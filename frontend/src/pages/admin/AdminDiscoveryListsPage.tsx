import { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Edit, Trash2, Eye } from 'lucide-react';
import { Container, Card, CardHeader, CardBody, Button, Spinner } from '../../components';
import { discoveryListApi } from '../../lib/discovery-list-api';
import type { DiscoveryListWithStats } from '../../types/discoveryList';
import { useToast } from '../../context/ToastContext';

// Constants
const DEFAULT_LIST_LIMIT = 100;
const DELETE_CONFIRMATION_TIMEOUT = 5000;

export function AdminDiscoveryListsPage() {
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const deleteConfirmTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  useEffect(() => {
    return () => {
      if (deleteConfirmTimerRef.current) clearTimeout(deleteConfirmTimerRef.current);
    };
  }, []);

  // Fetch all discovery lists
  const { data: lists, isLoading } = useQuery({
    queryKey: ['admin', 'discovery-lists'],
    queryFn: () => discoveryListApi.admin.listAllDiscoveryLists({ limit: DEFAULT_LIST_LIMIT }),
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (listId: string) => discoveryListApi.admin.deleteDiscoveryList(listId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-lists'] });
      showToast('Discovery list deleted successfully', 'success');
      setDeleteConfirm(null);
    },
    onError: () => {
      showToast('Failed to delete discovery list', 'error');
    },
  });

  // Toggle active status mutation
  const toggleActiveMutation = useMutation({
    mutationFn: ({ listId, isActive }: { listId: string; isActive: boolean }) =>
      discoveryListApi.admin.updateDiscoveryList(listId, { is_active: !isActive }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-lists'] });
      showToast('Discovery list status updated', 'success');
    },
    onError: () => {
      showToast('Failed to update discovery list status', 'error');
    },
  });

  const handleDelete = (listId: string) => {
    if (deleteConfirm === listId) {
      deleteMutation.mutate(listId);
    } else {
      setDeleteConfirm(listId);
      if (deleteConfirmTimerRef.current) clearTimeout(deleteConfirmTimerRef.current);
      deleteConfirmTimerRef.current = setTimeout(() => setDeleteConfirm(null), DELETE_CONFIRMATION_TIMEOUT);
    }
  };

  const handleToggleActive = (list: DiscoveryListWithStats) => {
    toggleActiveMutation.mutate({ listId: list.id, isActive: list.is_active });
  };

  if (isLoading) {
    return (
      <Container className="py-8 flex justify-center">
        <Spinner size="xl" />
      </Container>
    );
  }

  return (
    <Container className="py-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Curated Collections</h1>
          <p className="text-muted-foreground">
            Create and manage editorial clip collections
          </p>
        </div>
        <Button asChild>
          <Link to="/admin/discovery-lists/new">
            <Plus className="w-4 h-4 mr-2" />
            Create New List
          </Link>
        </Button>
      </div>

      <Card>
        <CardHeader>
          <h2 className="text-xl font-semibold">All Collections</h2>
        </CardHeader>
        <CardBody>
          {!lists || lists.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <p className="text-lg">No curated collections yet</p>
              <p className="text-sm mt-2">Create your first discovery list to get started</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b border-border">
                  <tr>
                    <th className="text-left py-3 px-4 font-semibold">Name</th>
                    <th className="text-left py-3 px-4 font-semibold">Clips</th>
                    <th className="text-left py-3 px-4 font-semibold">Followers</th>
                    <th className="text-left py-3 px-4 font-semibold">Status</th>
                    <th className="text-left py-3 px-4 font-semibold">Featured</th>
                    <th className="text-left py-3 px-4 font-semibold">Updated</th>
                    <th className="text-right py-3 px-4 font-semibold">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {lists.map((list) => (
                    <tr
                      key={list.id}
                      className="border-b border-border hover:bg-accent transition-colors"
                    >
                      <td className="py-3 px-4">
                        <div>
                          <div className="font-medium">{list.name}</div>
                          {list.description && (
                            <div className="text-sm text-muted-foreground line-clamp-1">
                              {list.description}
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="py-3 px-4">{list.clip_count}</td>
                      <td className="py-3 px-4">{list.follower_count}</td>
                      <td className="py-3 px-4">
                        <button
                          onClick={() => handleToggleActive(list)}
                          className={`px-2 py-1 rounded text-xs font-medium ${
                            list.is_active
                              ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                              : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                          }`}
                        >
                          {list.is_active ? 'Active' : 'Inactive'}
                        </button>
                      </td>
                      <td className="py-3 px-4">
                        {list.is_featured && (
                          <span className="px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                            Featured
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-sm text-muted-foreground">
                        {new Date(list.updated_at).toLocaleDateString()}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex justify-end gap-2">
                          <Button asChild variant="ghost" size="sm">
                            <Link to={`/discover/lists/${list.id}`} target="_blank" title="Preview" aria-label={`Preview ${list.name}`}>
                              <Eye className="w-4 h-4" />
                            </Link>
                          </Button>
                          <Button asChild variant="ghost" size="sm">
                            <Link to={`/admin/discovery-lists/${list.id}/edit`} title="Edit" aria-label={`Edit ${list.name}`}>
                              <Edit className="w-4 h-4" />
                            </Link>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(list.id)}
                            title={deleteConfirm === list.id ? 'Click again to confirm' : 'Delete'}
                            className={deleteConfirm === list.id ? 'text-red-600' : ''}
                          >
                            <Trash2 className="w-4 h-4" />
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
