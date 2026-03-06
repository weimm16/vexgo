import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { ArrowLeft, Save, Shield, Key, Bot } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { configApi } from '@/lib/api';
import type { CommentModerationConfig } from '@/types';

export function CommentConfigPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [config, setConfig] = useState<CommentModerationConfig>({
    id: '',
    enabled: false,
    modelProvider: '',
    apiKey: '',
    apiEndpoint: '',
    modelName: 'gpt-3.5-turbo',
    moderationPrompt: '请审核以下评论内容是否合规。如果评论包含违法不良信息、人身攻击、色情低俗等内容，请返回 \'REJECT\'；如果评论合规，请返回 \'APPROVE\'。只需返回结果，不要解释。\n\n评论内容：\n{{content}}',
    blockKeywords: '',
    autoApproveEnabled: true,
    minScoreThreshold: 0.5,
    createdAt: '',
    updatedAt: ''
  });

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const response = await configApi.getCommentModerationConfig();
      setConfig(response.data);
    } catch (error: any) {
      console.error('加载评论审核配置失败:', error);
      toast.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await configApi.updateCommentModerationConfig(config);
      toast.success('保存成功');
    } catch (error: any) {
      console.error('保存配置失败:', error);
      toast.error(error.response?.data?.error || '保存配置失败');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center py-8 text-muted-foreground">加载中...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/admin')}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">评论审核配置</h1>
            <p className="text-muted-foreground">配置AI评论审核规则和参数</p>
          </div>
        </div>
        <Button onClick={handleSave} disabled={saving}>
          <Save className="h-4 w-4 mr-2" />
          {saving ? '保存中...' : '保存配置'}
        </Button>
      </div>

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              基础设置
            </CardTitle>
            <CardDescription>
              配置评论审核的开关和基本参数
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>启用AI评论审核</Label>
                <p className="text-sm text-muted-foreground">
                  开启后，新评论将经过AI审核后才会显示
                </p>
              </div>
              <Switch
                checked={config.enabled}
                onCheckedChange={(checked) => setConfig({ ...config, enabled: checked })}
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>自动批准低风险评论</Label>
                <p className="text-sm text-muted-foreground">
                  开启后，风险分数低于阈值的评论将自动通过
                </p>
              </div>
              <Switch
                checked={config.autoApproveEnabled}
                onCheckedChange={(checked) => setConfig({ ...config, autoApproveEnabled: checked })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="minScore">最低分数阈值</Label>
              <Input
                id="minScore"
                type="number"
                min="0"
                max="1"
                step="0.1"
                value={config.minScoreThreshold}
                onChange={(e) => setConfig({ ...config, minScoreThreshold: parseFloat(e.target.value) })}
              />
              <p className="text-sm text-muted-foreground">
                低于此分数的评论将被拒绝（0-1之间）
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              API配置
            </CardTitle>
            <CardDescription>
              配置AI服务提供商的API密钥和端点
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="modelProvider">模型提供商</Label>
              <Select
                value={config.modelProvider}
                onValueChange={(value) => setConfig({ ...config, modelProvider: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择AI模型提供商" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="azure">Azure OpenAI</SelectItem>
                  <SelectItem value="anthropic">Anthropic</SelectItem>
                  <SelectItem value="custom">自定义API</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="modelName">模型名称</Label>
              <Input
                id="modelName"
                value={config.modelName}
                onChange={(e) => setConfig({ ...config, modelName: e.target.value })}
                placeholder="如: gpt-3.5-turbo"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="apiKey">API密钥</Label>
              <Input
                id="apiKey"
                type="password"
                value={config.apiKey}
                onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                placeholder="留空则不更新密钥"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="apiEndpoint">API端点（可选）</Label>
              <Input
                id="apiEndpoint"
                value={config.apiEndpoint}
                onChange={(e) => setConfig({ ...config, apiEndpoint: e.target.value })}
                placeholder="如使用代理或自定义端点"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bot className="h-5 w-5" />
              审核规则
            </CardTitle>
            <CardDescription>
              配置AI审核的提示词和关键词过滤
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="moderationPrompt">审核提示词</Label>
              <Textarea
                id="moderationPrompt"
                value={config.moderationPrompt}
                onChange={(e) => setConfig({ ...config, moderationPrompt: e.target.value })}
                rows={6}
                placeholder="输入AI审核时使用的提示词"
              />
              <p className="text-sm text-muted-foreground">
                可使用 {'{{content}}'} 占位符代表评论内容
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="blockKeywords">屏蔽关键词</Label>
              <Textarea
                id="blockKeywords"
                value={config.blockKeywords}
                onChange={(e) => setConfig({ ...config, blockKeywords: e.target.value })}
                rows={3}
                placeholder="输入需要屏蔽的关键词，多个关键词用逗号分隔"
              />
              <p className="text-sm text-muted-foreground">
                包含这些关键词的评论将被自动拒绝
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
