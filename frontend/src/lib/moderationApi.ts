import api from './api';
import type { Post, PostsResponse } from '@/types';

// 获取待审核文章列表
export const getPendingPosts = (params?: { 
  page?: number; 
  limit?: number; 
}) =>
  api.get<PostsResponse>('/moderation/pending', { params });

// 审核通过文章
export const approvePost = (id: string) =>
  api.put<{ message: string; post: Post }>(`/moderation/approve/${id}`);

// 拒绝文章
export const rejectPost = (id: string) =>
  api.put<{ message: string; post: Post }>(`/moderation/reject/${id}`);