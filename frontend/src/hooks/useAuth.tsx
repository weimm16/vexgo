import {
  createContext,
  useContext,
  useState,
  useEffect,
  type ReactNode,
} from "react";
import type { User } from "@/types";
import { authApi } from "@/lib/api";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (
    email: string,
    password: string,
    captchaData?: { id: string; token: string; x: number },
  ) => Promise<void>;
  loginWithToken: (token: string) => Promise<void>;
  register: (
    username: string,
    email: string,
    password: string,
    captchaData?: { id: string; token: string; x: number },
  ) => Promise<void>;
  logout: () => void;
  updateUser: (user: User) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    const savedUser = localStorage.getItem("user");

    if (token && savedUser) {
      setUser(JSON.parse(savedUser));
      authApi
        .getMe()
        .then((response) => {
          setUser(response.data.user);
          localStorage.setItem("user", JSON.stringify(response.data.user));
        })
        .catch(() => {
          localStorage.removeItem("token");
          localStorage.removeItem("user");
          setUser(null);
        })
        .finally(() => {
          setIsLoading(false);
        });
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = async (
    email: string,
    password: string,
    captchaData?: { id: string; token: string; x: number },
  ) => {
    const requestData: any = { email, password };
    if (captchaData) {
      requestData.captcha_id = captchaData.id;
      requestData.captcha_token = captchaData.token;
      requestData.captcha_x = captchaData.x;
    }
    const response = await authApi.login(requestData);
    const { user, token } = response.data;
    if (token) {
      localStorage.setItem("token", token);
      localStorage.setItem("user", JSON.stringify(user));
      setUser(user);
    }
  };

  // SSO 登录：直接接受后端签发的 JWT，拉取用户信息后更新 context
  const loginWithToken = async (token: string) => {
    localStorage.setItem("token", token);
    const response = await authApi.getMe();
    const u = response.data.user;
    localStorage.setItem("user", JSON.stringify(u));
    setUser(u);
  };

  const register = async (
    username: string,
    email: string,
    password: string,
    captchaData?: { id: string; token: string; x: number },
  ) => {
    const requestData: any = { username, email, password };
    if (captchaData) {
      requestData.captcha_id = captchaData.id;
      requestData.captcha_token = captchaData.token;
      requestData.captcha_x = captchaData.x;
    }
    try {
      const response = await authApi.register(requestData);
      const { user, token, requires_verification, email_verified } =
        response.data;

      if (requires_verification && !email_verified) {
        setUser(user);
        const error = new Error(response.data.message || "请先验证您的邮箱地址");
        (error as any).requiresVerification = true;
        (error as any).email = email;
        throw error;
      }

      if (token) {
        localStorage.setItem("token", token);
        localStorage.setItem("user", JSON.stringify(user));
        setUser(user);
      } else {
        setUser(user);
      }
    } catch (error: any) {
      // Extract error message from response
      const errorMessage = error.response?.data?.error || error.message || "注册失败";
      throw new Error(errorMessage);
    }
  };

  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setUser(null);
  };

  const updateUser = (updatedUser: User) => {
    setUser(updatedUser);
    localStorage.setItem("user", JSON.stringify(updatedUser));
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        loginWithToken,
        register,
        logout,
        updateUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}