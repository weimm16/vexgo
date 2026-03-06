import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { configApi } from '@/lib/api';
import type { GeneralSettings } from '@/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Settings, Save } from 'lucide-react';
import { toast } from 'sonner';

export function GeneralSettingsPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [config, setConfig] = useState<GeneralSettings>({
    id: '',
    captchaEnabled: false,
    registrationEnabled: true,
    siteName: 'VexGo',
    siteDescription: '',
    itemsPerPage: 20,
    createdAt: '',
    updatedAt: ''
  });

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const response = await configApi.getGeneralSettings();
      setConfig(response.data);
    } catch (error) {
      console.error('加载通用设置失败:', error);
      toast.error('加载设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!config.siteName.trim()) {
      toast.error('请输入网站名称');
      return;
    }
    if (config.itemsPerPage <= 0 || config.itemsPerPage > 100) {
      toast.error('每页显示数量必须在 1-100 之间');
      return;
    }

    setSaving(true);
    try {
      await configApi.updateGeneralSettings(config);
      toast.success('通用设置已保存');
    } catch (error) {
      console.error('保存通用设置失败:', error);
      const err = error as { response?: { data?: { error?: string } } };
      toast.error('保存失败: ' + (err.response?.data?.error || '未知错误'));
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        <div className="mb-6">
          <div className="h-8 w-48 bg-muted rounded animate-pulse mb-2" />
          <div className="h-4 w-64 bg-muted rounded animate-pulse" />
        </div>
        <Card>
          <CardContent className="p-6 space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-10 bg-muted rounded animate-pulse" />
            ))}
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      {/* 头部 */}
      <div className="mb-6">
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={() => navigate('/admin')}
          className="mb-4"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          返回管理后台
        </Button>
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Settings className="w-8 h-8" />
          通用设置
        </h1>
        <p className="text-muted-foreground mt-2">
          管理系统的基本配置，包括验证、注册、显示等设置
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>基本设置</CardTitle>
          <CardDescription>
            配置系统的基本参数
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* 网站名称 */}
          <div className="space-y-2">
            <Label htmlFor="siteName">网站名称</Label>
            <Input
              id="siteName"
              value={config.siteName}
              onChange={(e) => setConfig({ ...config, siteName: e.target.value })}
              placeholder="请输入网站名称"
            />
          </div>

          {/* 网站描述 */}
          <div className="space-y-2">
            <Label htmlFor="siteDescription">网站描述</Label>
            <Input
              id="siteDescription"
              value={config.siteDescription}
              onChange={(e) => setConfig({ ...config, siteDescription: e.target.value })}
              placeholder="请输入网站描述（可选）"
            />
          </div>

          {/* 每页显示数量 */}
          <div className="space-y-2">
            <Label htmlFor="itemsPerPage">每页显示数量</Label>
            <Input
              id="itemsPerPage"
              type="number"
              min={1}
              max={100}
              value={config.itemsPerPage}
              onChange={(e) => setConfig({ ...config, itemsPerPage: parseInt(e.target.value) || 20 })}
              placeholder="每页显示的项目数量"
            />
          </div>

          {/* 启用滑块验证 */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label htmlFor="captchaEnabled">滑块验证</Label>
              <p className="text-sm text-muted-foreground">
                启用滑动拼图验证码，增强安全性
              </p>
            </div>
            <Switch
              id="captchaEnabled"
              checked={config.captchaEnabled}
              onCheckedChange={(checked) => setConfig({ ...config, captchaEnabled: checked })}
            />
          </div>

          {/* 允许注册 */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label htmlFor="registrationEnabled">允许注册</Label>
              <p className="text-sm text-muted-foreground">
                允许新用户注册账号
              </p>
            </div>
            <Switch
              id="registrationEnabled"
              checked={config.registrationEnabled}
              onCheckedChange={(checked) => setConfig({ ...config, registrationEnabled: checked })}
            />
          </div>
        </CardContent>
      </Card>

      {/* 保存按钮 */}
      <div className="mt-6 flex justify-end">
        <Button onClick={handleSave} disabled={saving} size="lg">
          {saving ? (
            <>保存中...</>
          ) : (
            <>
              <Save className="w-4 h-4 mr-2" />
              保存设置
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
