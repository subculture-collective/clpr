import { apiClient } from './api';
import type { CreateReportRequest, UpdateReportRequest, Report, ReportWithDetails, ReportListResponse } from '../types/report';

// Submit a report
export const submitReport = async (data: CreateReportRequest): Promise<{ message: string; report: Report }> => {
  const response = await apiClient.post('/reports', data);
  return response.data;
};

// List reports (admin only)
export const listReports = async (
  page: number = 1,
  limit: number = 20,
  status?: string,
  type?: string
): Promise<ReportListResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });
  
  if (status) params.append('status', status);
  if (type) params.append('type', type);

  const response = await apiClient.get(`/admin/reports?${params}`);
  return response.data;
};

// Get report details (admin only)
export const getReport = async (reportId: string): Promise<ReportWithDetails> => {
  const response = await apiClient.get(`/admin/reports/${reportId}`);
  return response.data;
};

// Update report (admin only)
export const updateReport = async (reportId: string, data: UpdateReportRequest): Promise<{ message: string }> => {
  const response = await apiClient.put(`/admin/reports/${reportId}`, data);
  return response.data;
};
