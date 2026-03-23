import { useState, useEffect } from 'react';
import { Container, Card, Button, Badge, Spinner, Alert, Modal, TextArea } from '../../components';
import { Film, MessageSquare, User as UserIcon, FileText } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { listReports, updateReport } from '../../lib/report-api';
import type { Report, UpdateReportRequest } from '../../types/report';

export function AdminReportsPage() {
  const { isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();
  const [reports, setReports] = useState<Report[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState('pending');
  const [typeFilter, setTypeFilter] = useState('');

  // Action modal state
  const [actionModalOpen, setActionModalOpen] = useState(false);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [actionType, setActionType] = useState<string>('');
  const [actionNotes, setActionNotes] = useState('');

  const loadReports = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await listReports(page, 20, statusFilter, typeFilter);
      setReports(response.data ?? []);
      setTotalPages(response.meta?.total_pages ?? 1);
      setTotal(response.meta?.total ?? 0);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to load reports');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!isAuthenticated || !isAdmin) {
      navigate('/');
      return;
    }

    loadReports();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, isAdmin, navigate, page, statusFilter, typeFilter]);

  const openActionModal = (report: Report, action: string) => {
    setSelectedReport(report);
    setActionType(action);
    setActionModalOpen(true);
    setActionNotes('');
  };

  const handleAction = async () => {
    if (!selectedReport) return;

    try {
      const status = actionType === 'dismiss' ? 'dismissed' : 'actioned';
      const action = actionType === 'dismiss' ? 'mark_false' : actionType;

      await updateReport(selectedReport.id, {
        status,
        action: action as UpdateReportRequest['action'],
      });

      setSuccess(`Report ${status} successfully!`);
      setActionModalOpen(false);
      setSelectedReport(null);
      setActionType('');
      setActionNotes('');
      loadReports();
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to process report');
    }
  };

  const getReasonLabel = (reason: string) => {
    const labels: Record<string, string> = {
      spam: 'Spam',
      harassment: 'Harassment',
      nsfw: 'NSFW',
      violence: 'Violence',
      copyright: 'Copyright',
      other: 'Other',
    };
    return labels[reason] || reason;
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
      pending: 'warning',
      reviewed: 'default',
      actioned: 'success',
      dismissed: 'default',
    };
    return <Badge variant={variants[status] || 'default'}>{status}</Badge>;
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'clip':
        return <Film size={16} strokeWidth={1.75} />;
      case 'comment':
        return <MessageSquare size={16} strokeWidth={1.75} />;
      case 'user':
        return <UserIcon size={16} strokeWidth={1.75} />;
      default:
        return <FileText size={16} strokeWidth={1.75} />;
    }
  };

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <Container className="py-8">
      <h1 className="text-3xl font-bold mb-2">Report Management</h1>
      <p className="text-muted-foreground mb-6">Review user reports and take action</p>

      {error && (
        <Alert variant="error" className="mb-6" dismissible onDismiss={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert variant="success" className="mb-6" dismissible onDismiss={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      {/* Filters */}
      <Card className="p-4 mb-6">
        <div className="flex flex-wrap gap-4">
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium mb-2">Status</label>
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 bg-background border border-border rounded-lg"
            >
              <option value="">All</option>
              <option value="pending">Pending</option>
              <option value="reviewed">Reviewed</option>
              <option value="actioned">Actioned</option>
              <option value="dismissed">Dismissed</option>
            </select>
          </div>

          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium mb-2">Type</label>
            <select
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 bg-background border border-border rounded-lg"
            >
              <option value="">All</option>
              <option value="clip">Clips</option>
              <option value="comment">Comments</option>
              <option value="user">Users</option>
            </select>
          </div>

          <div className="flex items-end">
            <Button onClick={loadReports} variant="secondary" disabled={isLoading}>
              Refresh
            </Button>
          </div>
        </div>
      </Card>

      {/* Stats */}
      <Card className="p-4 mb-6">
        <div className="text-2xl font-bold">{total}</div>
        <div className="text-sm text-muted-foreground">Total Reports</div>
      </Card>

      {/* Reports List */}
      <Card className="p-6">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : reports.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">No reports found.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {reports.map((report) => (
              <div
                key={report.id}
                className="p-4 bg-background-secondary rounded-lg border border-border"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    {getTypeIcon(report.reportable_type)}
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium capitalize">{report.reportable_type}</span>
                        {getStatusBadge(report.status)}
                        <Badge variant="default">{getReasonLabel(report.reason)}</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        Reported {new Date(report.created_at).toLocaleString()}
                      </p>
                    </div>
                  </div>
                </div>

                {report.description && (
                  <p className="text-sm mb-3 p-3 bg-background-tertiary rounded">
                    {report.description}
                  </p>
                )}

                <div className="flex gap-2">
                  {report.status === 'pending' && (
                    <>
                      <Button
                        size="sm"
                        onClick={() => openActionModal(report, 'remove_content')}
                        className="bg-red-600 hover:bg-red-700"
                      >
                        Remove Content
                      </Button>
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => openActionModal(report, 'ban_user')}
                      >
                        Ban User
                      </Button>
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => openActionModal(report, 'dismiss')}
                      >
                        Dismiss
                      </Button>
                    </>
                  )}
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => navigate(`/admin/reports/${report.id}`)}
                  >
                    View Details
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-6">
            <Button
              variant="secondary"
              onClick={() => setPage(page - 1)}
              disabled={page === 1}
            >
              Previous
            </Button>
            <span className="flex items-center px-4">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="secondary"
              onClick={() => setPage(page + 1)}
              disabled={page === totalPages}
            >
              Next
            </Button>
          </div>
        )}
      </Card>

      {/* Action Modal */}
      {actionModalOpen && selectedReport && (
        <Modal
          open={actionModalOpen}
          onClose={() => {
            setActionModalOpen(false);
            setSelectedReport(null);
            setActionType('');
            setActionNotes('');
          }}
          title={`${actionType === 'dismiss' ? 'Dismiss' : 'Action'} Report`}
        >
          <div className="space-y-4">
            <p className="text-muted-foreground">
              {actionType === 'remove_content' && 'This will remove the reported content and mark the report as actioned.'}
              {actionType === 'ban_user' && 'This will ban the user who created the reported content.'}
              {actionType === 'dismiss' && 'This will mark the report as dismissed (false report).'}
            </p>

            <div>
              <label className="block text-sm font-medium mb-2">Notes (optional)</label>
              <TextArea
                value={actionNotes}
                onChange={(e) => setActionNotes(e.target.value)}
                placeholder="Add any notes about this action..."
                rows={3}
              />
            </div>

            <div className="flex gap-3">
              <Button
                onClick={handleAction}
                className={actionType === 'remove_content' || actionType === 'ban_user' ? 'flex-1 bg-red-600 hover:bg-red-700' : 'flex-1'}
              >
                Confirm
              </Button>
              <Button
                onClick={() => {
                  setActionModalOpen(false);
                  setSelectedReport(null);
                  setActionType('');
                  setActionNotes('');
                }}
                variant="secondary"
                className="flex-1"
              >
                Cancel
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </Container>
  );
}
