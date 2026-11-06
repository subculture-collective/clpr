import { apiClient } from './api';

export interface ContactMessageRequest {
  email: string;
  category: 'abuse' | 'account' | 'billing' | 'feedback';
  subject: string;
  message: string;
}

export interface ContactMessageResponse {
  message: string;
  status: string;
}

// Submit a contact message
export const submitContactMessage = async (data: ContactMessageRequest): Promise<ContactMessageResponse> => {
  const response = await apiClient.post('/contact', data);
  return response.data;
};
