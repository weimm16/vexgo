import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  CheckCircle, XCircle, Clock, User
} from 'lucide-react';
import { toast } from 'sonner';
import api from '@/lib/api';
import type { Comment } from '@/types';

interface PendingCommentsResponse {
  comments: Comment[];
  total: number;
}

interface ApprovedCommentsResponse {
  comments: Comment[];
  total: number;
}

interface RejectedCommentsResponse {
  comments: Comment[];
  total: number;
}

export function CommentModerationPage() {
  const [pendingComments, setPendingComments] = useState<Comment[]>([]);
  const [approvedComments, setApprovedComments] = useState<Comment[]>([]);
  const [rejectedComments, setRejectedComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('pending');

  useEffect(() => {
    loadData();
  }, [activeTab]);

  const loadData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'pending') {
        const response = await api.get<PendingCommentsResponse>('/moderation/comments/pending');
        setPendingComments(response.data.comments);
      } else if (activeTab === 'approved') {
        const response = await api.get<ApprovedCommentsResponse>('/moderation/comments/approved');
        setApprovedComments(response.data.comments);
      } else if (activeTab === 'rejected') {
        const response = await api.get<RejectedCommentsResponse>('/moderation/comments/rejected');
        setRejectedComments(response.data.comments);
      }
    } catch (error) {
      console.error('加载评论失败:', error);
      toast.error('加载评论失败');
    } finally {
      setLoading(false);
    }
  };

  const handleApproveComment = async (commentId: string) => {
    try {
      await api.put(`/moderation/comments/approve/${commentId}`);
      toast.success('评论审核通过');
      loadData();
    } catch (error) {
      console.error('审核通过失败:', error);
      toast.error('审核通过失败');
    }
  };

  const handleRejectComment = async (commentId: string) => {
    try {
      await api.put(`/moderation/comments/reject/${commentId}`);
      toast.success('评论已拒绝');
      loadData();
    } catch (error) {
      console.error('拒绝评论失败:', error);
      toast.error('拒绝评论失败');
    }
  };

  const getCurrentComments = () => {
    switch (activeTab) {
      case 'pending':
        return pendingComments;
      case 'approved':
        return approvedComments;
      case 'rejected':
        return rejectedComments;
      default:
        return [];
    }
  };

  const comments = getCurrentComments();

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">评论审核</h1>
          <p className="text-muted-foreground">管理用户提交的评论</p>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="mb-4">
          <TabsTrigger value="pending" className="flex items-center gap-2">
            <Clock className="h-4 w-4" />
            待审核
            <Badge variant="secondary" className="ml-1">
              {pendingComments.length}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="approved" className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4" />
            已通过
          </TabsTrigger>
          <TabsTrigger value="rejected" className="flex items-center gap-2">
            <XCircle className="h-4 w-4" />
            已拒绝
          </TabsTrigger>
        </TabsList>

        <TabsContent value="pending">
          <Card>
            <CardHeader>
              <CardTitle>待审核评论</CardTitle>
              <CardDescription>
                以下评论需要人工审核
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground">加载中...</div>
              ) : comments.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">暂无待审核评论</div>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border rounded-lg p-4">
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2 mb-2">
                          <User className="h-4 w-4" />
                          <span className="font-medium">{comment.author?.username || '匿名用户'}</span>
                          <span className="text-muted-foreground text-sm">
                            {new Date(comment.createdAt).toLocaleString('zh-CN')}
                          </span>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            size="sm"
                            variant="default"
                            onClick={() => handleApproveComment(comment.id)}
                          >
                            <CheckCircle className="h-4 w-4 mr-1" />
                            通过
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleRejectComment(comment.id)}
                          >
                            <XCircle className="h-4 w-4 mr-1" />
                            拒绝
                          </Button>
                        </div>
                      </div>
                      <div className="bg-muted/50 rounded p-3 mt-2">
                        <p className="text-sm">{comment.content}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="approved">
          <Card>
            <CardHeader>
              <CardTitle>已通过评论</CardTitle>
              <CardDescription>
                已审核通过并公开显示的评论
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground">加载中...</div>
              ) : comments.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">暂无已通过评论</div>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-2">
                        <User className="h-4 w-4" />
                        <span className="font-medium">{comment.author?.username || '匿名用户'}</span>
                        <span className="text-muted-foreground text-sm">
                          {new Date(comment.createdAt).toLocaleString('zh-CN')}
                        </span>
                        <Badge variant="default" className="ml-auto">已通过</Badge>
                      </div>
                      <div className="bg-muted/50 rounded p-3 mt-2">
                        <p className="text-sm">{comment.content}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="rejected">
          <Card>
            <CardHeader>
              <CardTitle>已拒绝评论</CardTitle>
              <CardDescription>
                已被拒绝的评论，不会公开显示
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground">加载中...</div>
              ) : comments.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">暂无已拒绝评论</div>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-2">
                        <User className="h-4 w-4" />
                        <span className="font-medium">{comment.author?.username || '匿名用户'}</span>
                        <span className="text-muted-foreground text-sm">
                          {new Date(comment.createdAt).toLocaleString('zh-CN')}
                        </span>
                        <Badge variant="destructive" className="ml-auto">已拒绝</Badge>
                      </div>
                      <div className="bg-muted/50 rounded p-3 mt-2">
                        <p className="text-sm">{comment.content}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}