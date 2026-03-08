import { useEffect, useState } from 'react';
import { useSearchParams, Link, useNavigate } from 'react-router-dom';
import { authApi } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, CheckCircle, XCircle } from 'lucide-react';

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const navigate = useNavigate();
  
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>(token ? 'loading' : 'error');
  const [message, setMessage] = useState(token ? '' : '验证令牌不能为空');
  const [requireRelogin, setRequireRelogin] = useState(false);

  useEffect(() => {
    if (!token) {
      return;
    }

    const verifyEmail = async () => {
      try {
        const response = await authApi.verifyEmail(token);
        setStatus('success');
        setMessage(response.data.message || '邮箱验证成功！');
        
        // 检查是否需要重新登录（邮箱变更成功）
        if (response.data.require_relogin) {
          setRequireRelogin(true);
          // 清除本地登录状态
          localStorage.removeItem('token');
          localStorage.removeItem('user');
        }
      } catch (err: unknown) {
        setStatus('error');
        const errorMessage = err instanceof Error ? err.message : '验证失败，请检查链接是否正确';
        setMessage(errorMessage);
      }
    };

    verifyEmail();
  }, [token]);

  return (
    <div className="container mx-auto px-4 py-16 flex justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl flex items-center justify-center gap-2">
            {status === 'loading' && <Loader2 className="w-6 h-6 animate-spin" />}
            {status === 'success' && <CheckCircle className="w-6 h-6 text-green-500" />}
            {status === 'error' && <XCircle className="w-6 h-6 text-red-500" />}
            邮箱验证
          </CardTitle>
          <CardDescription>
            {status === 'loading' && '正在验证您的邮箱...'}
            {status === 'success' && '验证完成'}
            {status === 'error' && '验证失败'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-muted-foreground">{message}</p>
          
          <div className="flex gap-3">
            {requireRelogin ? (
              <Button asChild className="flex-1" onClick={() => navigate('/login')}>
                前往登录
              </Button>
            ) : (
              <>
                <Button asChild className="flex-1">
                  <Link to="/login">
                    前往登录
                  </Link>
                </Button>
                {status === 'error' && (
                  <Button variant="outline" asChild className="flex-1">
                    <Link to="/">
                      返回首页
                    </Link>
                  </Button>
                )}
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
