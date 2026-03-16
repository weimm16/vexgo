import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import api from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Check, AlertCircle, Palette } from 'lucide-react';

interface ThemeInfo {
  id: string;
  name: string;
  author: string;
  version: string;
  description: string;
  url: string;
}

export function AdminThemePage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  
  const [themes, setThemes] = useState<ThemeInfo[]>([]);
  const [selectedTheme, setSelectedTheme] = useState<string>('default');
  const [loading, setLoading] = useState(true);
  const [applying, setApplying] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    // Check if user is admin or super admin
    if (user && user.role !== 'admin' && user.role !== 'super_admin') {
      navigate('/');
      return;
    }

    loadThemes();
    loadSelectedTheme();
  }, [user, navigate]);

  const loadThemes = async () => {
    setLoading(true);
    try {
      const response = await api.get<{ themes: ThemeInfo[] }>('/themes');
      setThemes(response.data.themes || []);
    } catch (error) {
      console.error('加载主题列表失败:', error);
      setMessage({ type: 'error', text: '加载主题列表失败' });
    } finally {
      setLoading(false);
    }
  };

  const loadSelectedTheme = () => {
    const theme = new URLSearchParams(window.location.search).get('theme') 
      || localStorage.getItem('selectedTheme') 
      || 'default';
    setSelectedTheme(theme);
  };

  const handleApplyTheme = async (themeId: string) => {
    setApplying(true);
    try {
      // Store in localStorage
      localStorage.setItem('selectedTheme', themeId);
      
      // Set cookie for server-side detection
      const expiryDate = new Date();
      expiryDate.setFullYear(expiryDate.getFullYear() + 1);
      document.cookie = `theme=${encodeURIComponent(themeId)}; expires=${expiryDate.toUTCString()}; path=/`;
      
      setSelectedTheme(themeId);
      
      const themeName = themes.find(t => t.id === themeId)?.name || themeId;
      setMessage({ type: 'success', text: `主题 "${themeName}" 已应用` });
      
      // Reload page to apply new theme
      setTimeout(() => {
        window.location.reload();
      }, 1500);
    } catch (error) {
      console.error('应用主题失败:', error);
      setMessage({ type: 'error', text: '应用主题失败' });
    } finally {
      setApplying(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Palette className="w-8 h-8" />
          主题管理
        </h1>
        <p className="text-gray-500 mt-2">选择和管理网站主题</p>
      </div>

      {message && (
        <Alert className={message.type === 'success' ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'}>
          {message.type === 'success' ? (
            <Check className="w-4 h-4 text-green-600" />
          ) : (
            <AlertCircle className="w-4 h-4 text-red-600" />
          )}
          <AlertDescription className={message.type === 'success' ? 'text-green-800' : 'text-red-800'}>
            {message.text}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4">
        {themes.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-center text-gray-500">未找到可用主题</p>
            </CardContent>
          </Card>
        ) : (
          themes.map((theme, index) => (
            <Card 
              key={index} 
              className={`cursor-pointer transition-all ${
                selectedTheme === theme.id 
                  ? 'border-blue-500 border-2 bg-blue-50' 
                  : 'hover:border-gray-400'
              }`}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <CardTitle className="flex items-center gap-2">
                      {theme.name}
                      {selectedTheme === theme.id && (
                        <Badge className="bg-blue-500">当前使用</Badge>
                      )}
                    </CardTitle>
                    <CardDescription>
                      <div className="mt-2 space-y-1 text-sm">
                        <p><span className="font-semibold">作者:</span> {theme.author}</p>
                        <p><span className="font-semibold">版本:</span> {theme.version}</p>
                        <p><span className="font-semibold">描述:</span> {theme.description}</p>
                        {theme.url && (
                          <p>
                            <span className="font-semibold">链接:</span>{' '}
                            <a 
                              href={theme.url} 
                              target="_blank" 
                              rel="noopener noreferrer"
                              className="text-blue-500 hover:underline"
                            >
                              {theme.url}
                            </a>
                          </p>
                        )}
                      </div>
                    </CardDescription>
                  </div>
                  <Button
                    onClick={() => handleApplyTheme(theme.id)}
                    disabled={selectedTheme === theme.id || applying}
                    className={selectedTheme === theme.id ? '' : ''}
                  >
                    {applying && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                    {selectedTheme === theme.id ? '已应用' : '应用主题'}
                  </Button>
                </div>
              </CardHeader>
            </Card>
          ))
        )}
      </div>

      <Card className="bg-blue-50 border-blue-200">
        <CardHeader>
          <CardTitle className="text-blue-900">主题安装说明</CardTitle>
        </CardHeader>
        <CardContent className="text-blue-800 space-y-2">
          <p>1. 在 <code className="bg-white px-2 py-1 rounded">./data/theme/</code> 目录下创建主题文件夹</p>
          <p>2. 在主题根目录创建 <code className="bg-white px-2 py-1 rounded">vexgo-theme.json</code> 元数据文件</p>
          <p>3. 在主题目录中放置 <code className="bg-white px-2 py-1 rounded">dist/</code> 文件夹（包含编译后的静态资源）</p>
          <p>4. 刷新此页面以查看新主题</p>
        </CardContent>
      </Card>
    </div>
  );
}
