import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { configApi } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { SliderCaptcha } from '@/components/ui/slider-captcha';
import { Loader2, Mail, Lock, User, Eye, EyeOff } from 'lucide-react';

export function RegisterPage() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [captchaData, setCaptchaData] = useState<{ id: string; token: string; x: number } | null>(null);
  const [isCaptchaModalOpen, setIsCaptchaModalOpen] = useState(false);
  const [captchaEnabled, setCaptchaEnabled] = useState(false);

  useEffect(() => {
    loadCaptchaSettings();
  }, []);

  const loadCaptchaSettings = async () => {
    try {
      const response = await configApi.getGeneralSettings();
      setCaptchaEnabled(response.data.captchaEnabled);
    } catch (error) {
      console.error('加载验证设置失败:', error);
      // 如果加载失败，默认不启用滑块验证
      setCaptchaEnabled(false);
    }
  };

  // 验证码验证成功回调
  const handleCaptchaSuccess = (data: { id: string; token: string; x: number }) => {
    setCaptchaData(data);
    // 验证码验证成功后，直接执行注册
    performRegister();
  };

  const performRegister = async () => {
    setLoading(true);

    try {
      // 如果启用了滑块验证，传递验证码数据
      if (captchaEnabled && captchaData) {
        await register(username, email, password, captchaData);
      } else {
        // 传递空的验证码数据（undefined）
        await register(username, email, password);
      }
      // 注册成功后关闭验证码弹窗
      setIsCaptchaModalOpen(false);
      // 注册成功且不需要验证，跳转到首页
      navigate('/');
    } catch (err) {
      // 如果需要邮箱验证，显示提示信息但不跳转
      const error = err as { requiresVerification?: boolean; response?: { data?: { message?: string } }; message?: string };
      if (error.requiresVerification) {
        setError(error.message || '请先验证您的邮箱地址才能登录');
      } else {
        setError(error.response?.data?.message || '注册失败，请重试');
      }
      
      // 注册失败后重置验证码状态
      setCaptchaData(null);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // 验证
    if (username.length < 3) {
      setError('用户名至少需要3个字符');
      return;
    }

    if (password.length < 6) {
      setError('密码至少需要6个字符');
      return;
    }

    if (password !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }
    
    // 如果启用了滑块验证，打开验证码弹窗
    if (captchaEnabled) {
      // 如果已经有验证码数据，直接执行注册
      if (captchaData) {
        performRegister();
        return;
      }
      // 打开验证码弹窗
      setIsCaptchaModalOpen(true);
      return;
    }
    
    // 未启用滑块验证，直接执行注册
    performRegister();
  };

  return (
    <div className="container mx-auto px-4 py-16 flex justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">创建账号</CardTitle>
          <CardDescription>
            加入 BlogHub，开始你的创作之旅
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

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
                  required
                  minLength={3}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="your@email.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-10"
                  required
                />
              </div>
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
                  required
                  minLength={6}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
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
                  required
                />
              </div>
            </div>

            <Button
              type="submit"
              className="w-full mt-4"
              disabled={loading}
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
        onClose={() => setIsCaptchaModalOpen(false)}
        onSuccess={handleCaptchaSuccess}
      />
    </div>
  );
}