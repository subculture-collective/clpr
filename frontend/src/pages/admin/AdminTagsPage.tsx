import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2 } from 'lucide-react';
import { Container, Card, CardHeader, CardBody, Button, Spinner, SEO } from '../../components';
import { Input } from '../../components/ui';
import { apiClient } from '@/lib/api';
import { useToast } from '../../context/ToastContext';

interface BlacklistEntry {
  id: string;
  pattern: string;
  reason: string;
  created_at: string;
}

export function AdminTagsPage() {
  const [pattern, setPattern] = useState('');
  const [reason, setReason] = useState('');
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  // Fetch blacklisted tag patterns
  const { data: blacklist, isLoading } = useQuery<BlacklistEntry[]>({
    queryKey: ['admin', 'tags', 'blacklist'],
    queryFn: async () => {
      const res = await apiClient.get('/api/v1/admin/tags/blacklist');
      return res.data;
    },
  });

  // Add pattern mutation
  const addMutation = useMutation({
    mutationFn: async (payload: { pattern: string; reason: string }) => {
      const res = await apiClient.post('/api/v1/admin/tags/blacklist', payload);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags', 'blacklist'] });
      showToast('Blacklist pattern added successfully', 'success');
      setPattern('');
      setReason('');
    },
    onError: () => {
      showToast('Failed to add blacklist pattern', 'error');
    },
  });

  // Delete pattern mutation
  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/api/v1/admin/tags/blacklist/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags', 'blacklist'] });
      showToast('Blacklist pattern removed successfully', 'success');
    },
    onError: () => {
      showToast('Failed to remove blacklist pattern', 'error');
    },
  });

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedPattern = pattern.trim();
    if (!trimmedPattern) return;
    addMutation.mutate({ pattern: trimmedPattern, reason: reason.trim() });
  };

  if (isLoading) {
    return (
      <Container className='py-8 flex justify-center'>
        <Spinner size='xl' />
      </Container>
    );
  }

  return (
    <Container className='py-4 xs:py-6 md:py-8'>
      <SEO title='Tag Blacklist' noindex />

      <div className='mb-6 xs:mb-8'>
        <h1 className='text-2xl xs:text-3xl font-bold text-text-primary mb-2'>
          Tag Blacklist
        </h1>
        <p className='text-sm xs:text-base text-text-secondary'>
          Manage blacklisted tag patterns. Clips with tags matching these patterns will be filtered out.
        </p>
      </div>

      {/* Add Pattern Form */}
      <Card className='mb-6'>
        <CardHeader>
          <h2 className='text-xl font-semibold text-text-primary'>Add Pattern</h2>
        </CardHeader>
        <CardBody>
          <form onSubmit={handleAdd} className='flex flex-col sm:flex-row gap-3'>
            <Input
              placeholder='Pattern (e.g. spam*)'
              value={pattern}
              onChange={e => setPattern(e.target.value)}
              fullWidth
              className='sm:max-w-xs'
            />
            <Input
              placeholder='Reason (optional)'
              value={reason}
              onChange={e => setReason(e.target.value)}
              fullWidth
              className='sm:max-w-xs'
            />
            <Button
              type='submit'
              variant='primary'
              disabled={!pattern.trim() || addMutation.isPending}
              className='cursor-pointer self-start sm:self-end'
            >
              {addMutation.isPending ? <Spinner size='sm' /> : 'Add'}
            </Button>
          </form>
        </CardBody>
      </Card>

      {/* Blacklist Table */}
      <Card>
        <CardHeader>
          <h2 className='text-xl font-semibold text-text-primary'>Blacklisted Patterns</h2>
        </CardHeader>
        <CardBody>
          {!blacklist || blacklist.length === 0 ? (
            <div className='text-center py-12 text-text-secondary'>
              <p className='text-lg'>No blacklisted patterns yet</p>
              <p className='text-sm mt-2'>Add a pattern above to get started</p>
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <table className='w-full' role='table' aria-label='Blacklisted tag patterns'>
                <thead className='border-b border-border'>
                  <tr>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Pattern</th>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Reason</th>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Added</th>
                    <th className='text-right py-3 px-4 font-semibold text-text-primary'>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {blacklist.map(entry => (
                    <tr
                      key={entry.id}
                      className='border-b border-border hover:bg-surface transition-colors'
                    >
                      <td className='py-3 px-4'>
                        <code className='text-sm bg-surface px-2 py-0.5 rounded'>
                          {entry.pattern}
                        </code>
                      </td>
                      <td className='py-3 px-4 text-sm text-text-secondary'>
                        {entry.reason || '-'}
                      </td>
                      <td className='py-3 px-4 text-sm text-text-secondary'>
                        {new Date(entry.created_at).toLocaleDateString()}
                      </td>
                      <td className='py-3 px-4'>
                        <div className='flex justify-end'>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => deleteMutation.mutate(entry.id)}
                            disabled={deleteMutation.isPending}
                            className='cursor-pointer text-red-600 hover:text-red-700'
                            title='Delete pattern'
                            aria-label={`Delete pattern ${entry.pattern}`}
                          >
                            <Trash2 className='w-4 h-4' />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardBody>
      </Card>
    </Container>
  );
}

export default AdminTagsPage;
