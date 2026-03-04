import api from './api';
import type { User, Pagination } from '@/types';

// 用户列表响应类型
export interface UsersResponse {
  users: User[];
  pagination: Pagination;
}

// 获取用户列表
export const getUsers = (params?: { 
  page?: number; 
  limit?: number; 
}) =>
  api.get<UsersResponse>('/users', { params });

// 更新用户角色
export const updateUserRole = (id: string, role: string) =>
  api.put<{ message: string; user: User }>(`/users/${id}/role`, { role });