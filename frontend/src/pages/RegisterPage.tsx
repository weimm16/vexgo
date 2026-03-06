import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { configApi } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { SliderCaptcha } from '@/components/ui/slider-captcha';
import { Loader2, Mail, Lock, User, Eye, EyeOff, CheckCircle } from 'lucide-react';
import { toast } from 'sonner';

export function RegisterPage() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [usernameError, setUsernameError] = useState('');
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [confirmPasswordError, setConfirmPasswordError] = useState('');
  const [captchaData, setCaptchaData] = useState<{ id: string; token: string; x: number } | null>(null);
  const [isCaptchaModalOpen, setIsCaptchaModalOpen] = useState(false);
  const [captchaEnabled, setCaptchaEnabled] = useState(false);
  const [isCaptchaVerified, setIsCaptchaVerified] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState(true);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const response = await configApi.getGeneralSettings();
      setCaptchaEnabled(response.data.captchaEnabled);
      setRegistrationEnabled(response.data.registrationEnabled);
    } catch (error) {
      console.error('加载设置失败:', error);
    }
  };

  // 验证码验证成功回调
  const handleCaptchaSuccess = async (data: { id: string; token: string; x: number }) => {
    setCaptchaData(data);
    
    // 调用预验证接口，标记验证码为已使用
    try {
      await fetch('/api/captcha/verify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id: data.id,
          token: data.token,
          x: data.x,
        }),
      });
    } catch (error) {
      console.error('预验证失败:', error);
      // 即使预验证失败，也保存数据，让用户尝试注册
    }
    
    setIsCaptchaVerified(true);
    // 不自动执行注册，等待用户点击注册按钮
  };

  // 重置验证码状态
  const resetCaptcha = () => {
    setCaptchaData(null);
    setIsCaptchaVerified(false);
  };

  const performRegister = async () => {
    setLoading(true);
    try {
      if (captchaEnabled) {
        if (!captchaData) {
          toast.error('请先完成滑块验证');
          setIsCaptchaModalOpen(true);
          return;
        }
        await register(username, email, password, captchaData);
      } else {
        await register(username, email, password);
      }
      setIsCaptchaModalOpen(false);
      resetCaptcha();
      navigate('/');
    } catch (err) {
      // 邮箱验证等后端错误仍可弹窗提示
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      const msg = error.response?.data?.message || error.message || '';
      toast.error(msg || '注册失败，请重试');
      // 注册失败后不重置验证码状态，允许用户重试
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setUsernameError('');
    setEmailError('');
    setPasswordError('');
    setConfirmPasswordError('');
    let hasError = false;
    if (username.length < 3) {
      setUsernameError('用户名至少需要3个字符');
      hasError = true;
    }
    if (!email || !/^\S+@\S+\.\S+$/.test(email)) {
      setEmailError('请输入有效的邮箱地址');
      hasError = true;
    }
    if (password.length < 6) {
      setPasswordError('密码至少需要6个字符');
      hasError = true;
    }
    if (password !== confirmPassword) {
      setConfirmPasswordError('两次输入的密码不一致');
      hasError = true;
    }
    if (hasError) return;
    
    // 检查是否允许注册
    if (!registrationEnabled) {
      toast.error('系统已关闭注册功能，请联系管理员');
      return;
    }
    
    // 如果启用了滑块验证，检查是否已完成验证
    if (captchaEnabled) {
      if (!isCaptchaVerified || !captchaData) {
        toast.error('请先完成滑块验证');
        setIsCaptchaModalOpen(true);
        return;
      }
    }
    
    performRegister();
  };

  return (
    <div className="container mx-auto px-4 py-16 flex justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">创建账号</CardTitle>
          <CardDescription>
            加入 VexGo，开始你的创作之旅
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4" noValidate>
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="username"
                  type="text"
                  placeholder="请输入用户名"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="pl-10"
                />
              </div>
              {usernameError && <div className="text-red-500 text-sm mt-1">{usernameError}</div>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="请输入邮箱"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-10"
                />
              </div>
              {emailError && <div className="text-red-500 text-sm mt-1">{emailError}</div>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="至少6个字符"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="pl-10 pr-10"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              {passwordError && <div className="text-red-500 text-sm mt-1">{passwordError}</div>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirmPassword">确认密码</Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="confirmPassword"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="再次输入密码"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="pl-10"
                />
              </div>
              {confirmPasswordError && <div className="text-red-500 text-sm mt-1">{confirmPasswordError}</div>}
            </div>

            {/* 滑块验证区域 */}
            {captchaEnabled && (
              <div className="mt-4">
                <div className="flex items-center justify-between mb-2">
                  <Label className="text-sm font-medium">安全验证</Label>
                  {isCaptchaVerified && (
                    <span className="text-xs text-green-600 flex items-center">
                      <CheckCircle className="h-3 w-3 mr-1" />
                      已验证
                    </span>
                  )}
                </div>
                <div className="border rounded-lg p-3 bg-gray-50">
                  {isCaptchaVerified ? (
                    <div className="flex items-center justify-between">
                      <div className="flex items-center text-sm text-green-600">
                        <CheckCircle className="h-4 w-4 mr-2" />
                        验证已完成
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={resetCaptcha}
                        className="text-xs h-7"
                      >
                        重新验证
                      </Button>
                    </div>
                  ) : (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">请完成滑块验证</span>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => setIsCaptchaModalOpen(true)}
                        className="text-xs h-7"
                      >
                        去验证
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}

            <Button
              type="submit"
              className="w-full mt-4"
              disabled={loading || (captchaEnabled && !isCaptchaVerified)}
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  注册中...
                </>
              ) : (
                '注册'
              )}
            </Button>
          </form>

          <div className="mt-6 text-center text-sm">
            <span className="text-muted-foreground">已有账号？</span>{' '}
            <Link to="/login" className="text-primary hover:underline">
              立即登录
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* 验证码弹窗 */}
      <SliderCaptcha
        isOpen={isCaptchaModalOpen}
        onClose={() => {
          // 允许用户随时关闭验证窗口
          setIsCaptchaModalOpen(false);
        }}
        onSuccess={handleCaptchaSuccess}
      />
    </div>
  );
}
