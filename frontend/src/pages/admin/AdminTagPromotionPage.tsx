import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, XCircle } from 'lucide-react';
import { Container, Card, CardHeader, CardBody, Button, Spinner, SEO } from '../../components';
import { tagApi } from '@/lib/tag-api';
import { useToast } from '../../context/ToastContext';
import type { TagPromotionItem } from '../../types/tag';

export function AdminTagPromotionPage() {
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  const { data: queue, isLoading, isError } = useQuery({
    queryKey: ['admin', 'tag-promotion'],
    queryFn: () => tagApi.getPromotionQueue({ status: 'pending' }),
  });

  const approveMutation = useMutation({
    mutationFn: (id: string) => tagApi.approvePromotion(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tag-promotion'] });
      showToast('Tag promoted to AI pool', 'success');
    },
    onError: () => {
      showToast('Failed to approve tag promotion', 'error');
    },
  });

  const rejectMutation = useMutation({
    mutationFn: (id: string) => tagApi.rejectPromotion(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tag-promotion'] });
      showToast('Tag promotion rejected', 'success');
    },
    onError: () => {
      showToast('Failed to reject tag promotion', 'error');
    },
  });

  if (isLoading) {
    return (
      <Container className='py-8 flex justify-center'>
        <Spinner size='xl' />
      </Container>
    );
  }

  if (isError) {
    return (
      <Container className='py-8'>
        <div className='text-center py-12 text-text-secondary'>
          <p className='text-lg'>Failed to load promotion queue</p>
          <p className='text-sm mt-2'>Please try again later</p>
        </div>
      </Container>
    );
  }

  const items = queue?.items ?? [];

  return (
    <Container className='py-4 xs:py-6 md:py-8'>
      <SEO title='Tag Promotion Queue' noindex />

      <div className='mb-6 xs:mb-8'>
        <h1 className='text-2xl xs:text-3xl font-bold text-text-primary mb-2'>
          Tag Promotion Queue
        </h1>
        <p className='text-sm xs:text-base text-text-secondary'>
          Review community-submitted tags that have crossed the usage threshold. Approved tags become available for AI assignment.
        </p>
      </div>

      <Card>
        <CardHeader>
          <h2 className='text-xl font-semibold text-text-primary'>
            Pending Promotions ({items.length})
          </h2>
        </CardHeader>
        <CardBody>
          {items.length === 0 ? (
            <div className='text-center py-12 text-text-secondary'>
              <p className='text-lg'>No pending promotions</p>
              <p className='text-sm mt-2'>
                Community tags will appear here once they reach the usage threshold (3+ unique users, 5+ clips).
              </p>
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <table className='w-full' role='table' aria-label='Tag promotion queue'>
                <thead className='border-b border-border'>
                  <tr>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Tag</th>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Unique Users</th>
                    <th className='text-left py-3 px-4 font-semibold text-text-primary'>Clip Count</th>
                    <th className='text-right py-3 px-4 font-semibold text-text-primary'>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((item: TagPromotionItem) => (
                    <tr
                      key={item.id}
                      className='border-b border-border hover:bg-surface transition-colors'
                    >
                      <td className='py-3 px-4'>
                        <code className='text-sm bg-surface px-2 py-0.5 rounded font-medium'>
                          #{item.tag_slug}
                        </code>
                      </td>
                      <td className='py-3 px-4 text-sm text-text-secondary'>
                        {item.unique_users}
                      </td>
                      <td className='py-3 px-4 text-sm text-text-secondary'>
                        {item.usage_count}
                      </td>
                      <td className='py-3 px-4'>
                        <div className='flex justify-end gap-2'>
                          <Button
                            variant='primary'
                            size='sm'
                            onClick={() => approveMutation.mutate(item.id)}
                            disabled={approveMutation.isPending || rejectMutation.isPending}
                            className='cursor-pointer'
                            title='Promote to AI pool'
                            aria-label={`Approve tag ${item.tag_slug}`}
                            leftIcon={<CheckCircle className='w-4 h-4' />}
                          >
                            Approve
                          </Button>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => rejectMutation.mutate(item.id)}
                            disabled={approveMutation.isPending || rejectMutation.isPending}
                            className='cursor-pointer text-red-600 hover:text-red-700'
                            title='Reject promotion'
                            aria-label={`Reject tag ${item.tag_slug}`}
                            leftIcon={<XCircle className='w-4 h-4' />}
                          >
                            Reject
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

export default AdminTagPromotionPage;
