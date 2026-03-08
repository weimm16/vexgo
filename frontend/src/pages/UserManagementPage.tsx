import { useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { getUsers, updateUserRole } from '@/lib/userApi';
import type { User } from '@/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/ui/select';
import { toast } from 'sonner';
import {
  Users, UserCheck
} from 'lucide-react';

export function UserManagementPage() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  // 角色显示映射
  const roleDisplayMap: Record<string, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
    'super_admin': { label: '超级管理员', variant: 'destructive' },
    'admin': { label: '管理员', variant: 'default' },
    'author': { label: '作者', variant: 'secondary' },
    'contributor': { label: '投稿者', variant: 'outline' },
    'guest': { label: '访客', variant: 'outline' }
  };

  // 可分配的角色选项（根据当前用户角色确定）
  const getAssignableRoles = () => {
    if (currentUser?.role === 'super_admin') {
      return [
        { value: 'admin', label: '管理员' },
        { value: 'author', label: '作者' },
        { value: 'contributor', label: '投稿者' },
        { value: 'guest', label: '访客' }
      ];
    } else if (currentUser?.role === 'admin') {
      return [
        { value: 'author', label: '作者' },
        { value: 'contributor', label: '投稿者' },
        { value: 'guest', label: '访客' }
      ];
    }
    return [];
  };

  useEffect(() => {
    loadData();
  }, [currentPage]);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await getUsers({ page: currentPage, limit: 10 });
      setUsers(response.data.users);
      setTotalPages(response.data.pagination.totalPages);
    } catch (error) {
      console.error('加载用户列表失败:', error);
      toast.error('加载用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      const response = await updateUserRole(userId, newRole);
      toast.success(response.data.message);
      
      // 更新本地用户列表
      setUsers(prevUsers =>
        prevUsers.map(user =>
          user.id === userId ? { ...user, role: newRole as 'super_admin' | 'admin' | 'author' | 'contributor' | 'guest' } : user
        )
      );
    } catch (error) {
      console.error('更新用户角色失败:', error);
      toast.error('更新用户角色失败');
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Users className="w-6 h-6" />
            用户管理
          </h1>
        </div>
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="animate-pulse space-y-4">
                  <div className="h-6 bg-muted rounded w-3/4"></div>
                  <div className="h-4 bg-muted rounded w-1/2"></div>
                  <div className="h-4 bg-muted rounded w-1/4"></div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Users className="w-6 h-6" />
          用户管理
        </h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <UserCheck className="w-5 h-5" />
            用户列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {users.map((user) => (
              <div 
                key={user.id} 
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                      <span className="text-primary font-medium">
                        {user.username.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <h3 className="font-medium">{user.username}</h3>
                      <p className="text-sm text-muted-foreground">{user.email}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span>注册时间: {user.createdAt ? formatDate(user.createdAt) : '未知'}</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-4">
                  <Badge variant={roleDisplayMap[user.role]?.variant || 'outline'}>
                    {roleDisplayMap[user.role]?.label || user.role}
                  </Badge>
                  
                  {currentUser?.id !== user.id && !(currentUser?.role === 'admin' && user.role === 'admin') && (
                    <div className="flex items-center gap-2">
                      <Select
                        value={user.role}
                        onValueChange={(value) => handleRoleChange(user.id, value)}
                      >
                        <SelectTrigger className="w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {getAssignableRoles().map((role) => (
                            <SelectItem key={role.value} value={role.value}>
                              {role.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                  
                  {currentUser?.id !== user.id && currentUser?.role === 'admin' && user.role === 'admin' && (
                    <Badge variant="secondary">同等级</Badge>
                  )}
                  
                  {currentUser?.id === user.id && (
                    <Badge variant="secondary">当前用户</Badge>
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-6">
              <div className="text-sm text-muted-foreground">
                第 {currentPage} 页，共 {totalPages} 页
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                  disabled={currentPage === totalPages}
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}