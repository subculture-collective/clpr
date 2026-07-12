import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import {
  listCampaigns,
  createCampaign,
  updateCampaign,
  deleteCampaign,
  getCTRReportByCampaign,
  getCampaignReportByPlacement,
  formatCurrency,
  formatPercentage,
  getCampaignStatus,
  getCampaignStatusColor,
  STANDARD_BANNER_SIZES,
} from '../../lib/campaigns-api';
import type {
  Campaign,
  CreateCampaignRequest,
  UpdateCampaignRequest,
} from '../../lib/campaigns-api';
import {
  MetricCard,
  BarChartComponent,
} from '../../components/analytics';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';

type TabType = 'campaigns' | 'reports';
type ModalType = 'create' | 'edit' | 'delete' | null;

const AdminCampaignsPage: React.FC = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<TabType>('campaigns');
  const [modalType, setModalType] = useState<ModalType>(null);
  const [selectedCampaign, setSelectedCampaign] = useState<Campaign | null>(null);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [reportDays, setReportDays] = useState(30);

  // Fetch campaigns
  const {
    data: campaignsData,
    isLoading: isLoadingCampaigns,
    error: campaignsError,
  } = useQuery({
    queryKey: ['adminCampaigns', page, statusFilter],
    queryFn: () => listCampaigns(page, 20, statusFilter || undefined),
    staleTime: 30 * 1000,
  });

  // Fetch CTR report by campaign
  const { data: ctrReport, isLoading: isLoadingCTR } = useQuery({
    queryKey: ['adminCampaignsCTR', reportDays],
    queryFn: () => getCTRReportByCampaign(reportDays),
    enabled: activeTab === 'reports',
    staleTime: 5 * 60 * 1000,
  });

  // Fetch placement report
  const { data: placementReport, isLoading: isLoadingPlacement } = useQuery({
    queryKey: ['adminCampaignsPlacement', reportDays],
    queryFn: () => getCampaignReportByPlacement(undefined, reportDays),
    enabled: activeTab === 'reports',
    staleTime: 5 * 60 * 1000,
  });

  // Mutations
  const createMutation = useMutation({
    mutationFn: createCampaign,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminCampaigns'] });
      setModalType(null);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateCampaignRequest }) =>
      updateCampaign(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminCampaigns'] });
      setModalType(null);
      setSelectedCampaign(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteCampaign,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminCampaigns'] });
      setModalType(null);
      setSelectedCampaign(null);
    },
  });

  // Calculate summary metrics
  const totalImpressions = ctrReport?.reports?.reduce((sum, r) => sum + r.impressions, 0) || 0;
  const totalClicks = ctrReport?.reports?.reduce((sum, r) => sum + r.clicks, 0) || 0;
  const totalSpend = ctrReport?.reports?.reduce((sum, r) => sum + r.spend_cents, 0) || 0;
  const avgCTR = totalImpressions > 0 ? (totalClicks / totalImpressions) * 100 : 0;

  const handleCreate = () => {
    setSelectedCampaign(null);
    setModalType('create');
  };

  const handleEdit = (campaign: Campaign) => {
    setSelectedCampaign(campaign);
    setModalType('edit');
  };

  const handleDelete = (campaign: Campaign) => {
    setSelectedCampaign(campaign);
    setModalType('delete');
  };

  const handleCloseModal = () => {
    setModalType(null);
    setSelectedCampaign(null);
  };

  return (
    <>
      <Helmet>
        <title>Ad Campaigns - Admin</title>
      </Helmet>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Ad Campaigns
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Manage ad campaigns, creatives, and view performance reports
          </p>
        </div>

        {/* Tabs */}
        <div className="border-b border-gray-200 dark:border-gray-700 mb-6">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('campaigns')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'campaigns'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
              }`}
            >
              Campaigns
            </button>
            <button
              onClick={() => setActiveTab('reports')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'reports'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
              }`}
            >
              Reports
            </button>
          </nav>
        </div>

        {/* Campaigns Tab */}
        {activeTab === 'campaigns' && (
          <div>
            {/* Actions Bar */}
            <div className="flex justify-between items-center mb-6">
              <div className="flex items-center space-x-4">
                <select
                  value={statusFilter}
                  onChange={(e) => {
                    setStatusFilter(e.target.value);
                    setPage(1);
                  }}
                  className="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-sm"
                >
                  <option value="">All Statuses</option>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                  <option value="scheduled">Scheduled</option>
                  <option value="ended">Ended</option>
                </select>
              </div>
              <button
                onClick={handleCreate}
                className="bg-primary text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-primary/90"
              >
                Create Campaign
              </button>
            </div>

            {/* Campaigns List */}
            {campaignsError && (
              <div className="bg-red-100 dark:bg-red-900 border border-red-400 text-red-700 dark:text-red-200 px-4 py-3 rounded mb-4">
                <p>Error loading campaigns. Please try again.</p>
              </div>
            )}

            {isLoadingCampaigns ? (
              <div className="animate-pulse space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
                  </div>
                ))}
              </div>
            ) : (
              <>
                <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
                  <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-900">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Campaign
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Type
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Status
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Budget
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Spent
                        </th>
                        <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                      {campaignsData?.campaigns?.length === 0 && (
                        <tr>
                          <td colSpan={6} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                            No campaigns found. Create your first campaign to get started.
                          </td>
                        </tr>
                      )}
                      {campaignsData?.campaigns?.map((campaign) => {
                        const status = getCampaignStatus(campaign);
                        return (
                          <tr key={campaign.id}>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <div className="text-sm font-medium text-gray-900 dark:text-white">
                                {campaign.name}
                              </div>
                              <div className="text-sm text-gray-500 dark:text-gray-400">
                                {campaign.advertiser_name}
                              </div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span className="text-sm text-gray-900 dark:text-white capitalize">
                                {campaign.ad_type}
                              </span>
                              {campaign.width && campaign.height && (
                                <div className="text-xs text-gray-500 dark:text-gray-400">
                                  {campaign.width}x{campaign.height}
                                </div>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span
                                className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getCampaignStatusColor(
                                  status
                                )}`}
                              >
                                {status}
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                              {campaign.daily_budget_cents
                                ? `${formatCurrency(campaign.daily_budget_cents)}/day`
                                : campaign.total_budget_cents
                                ? formatCurrency(campaign.total_budget_cents)
                                : 'Unlimited'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                              {formatCurrency(campaign.spent_total_cents)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                              <button
                                onClick={() => handleEdit(campaign)}
                                className="text-primary hover:text-primary/80 mr-4"
                              >
                                Edit
                              </button>
                              <button
                                onClick={() => handleDelete(campaign)}
                                className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                              >
                                Delete
                              </button>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Pagination */}
                {campaignsData && campaignsData.total_pages > 1 && (
                  <div className="flex justify-between items-center mt-4">
                    <p className="text-sm text-gray-700 dark:text-gray-400">
                      Showing page {page} of {campaignsData.total_pages} ({campaignsData.total}{' '}
                      total)
                    </p>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => setPage((p) => Math.max(1, p - 1))}
                        disabled={!campaignsData.has_prev}
                        className="px-3 py-1 rounded-md bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 disabled:opacity-50"
                      >
                        Previous
                      </button>
                      <button
                        onClick={() => setPage((p) => p + 1)}
                        disabled={!campaignsData.has_next}
                        className="px-3 py-1 rounded-md bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 disabled:opacity-50"
                      >
                        Next
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        )}

        {/* Reports Tab */}
        {activeTab === 'reports' && (
          <div>
            {/* Date Range Selector */}
            <div className="mb-6">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mr-2">
                Time Range:
              </label>
              <select
                value={reportDays}
                onChange={(e) => setReportDays(Number(e.target.value))}
                className="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-sm"
              >
                <option value={7}>Last 7 days</option>
                <option value={14}>Last 14 days</option>
                <option value={30}>Last 30 days</option>
                <option value={60}>Last 60 days</option>
                <option value={90}>Last 90 days</option>
              </select>
            </div>

            {/* Summary Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <MetricCard
                title="Total Impressions"
                value={totalImpressions.toLocaleString()}
                subtitle={`Last ${reportDays} days`}
              />
              <MetricCard
                title="Total Clicks"
                value={totalClicks.toLocaleString()}
                subtitle={`Last ${reportDays} days`}
              />
              <MetricCard
                title="Average CTR"
                value={formatPercentage(avgCTR)}
                subtitle="Click-through rate"
              />
              <MetricCard
                title="Total Spend"
                value={formatCurrency(totalSpend)}
                subtitle={`Last ${reportDays} days`}
              />
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
              {/* CTR by Campaign */}
              {isLoadingCTR ? (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 h-80 animate-pulse">
                  <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                  <div className="h-full bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              ) : ctrReport?.reports && ctrReport.reports.length > 0 ? (
                <BarChartComponent
                  data={ctrReport.reports.slice(0, 10).map((r) => ({
                    name: r.ad_name,
                    value: r.ctr,
                  }))}
                  title="CTR by Campaign"
                  valueLabel="CTR (%)"
                  color="#10b981"
                  height={350}
                />
              ) : (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 flex items-center justify-center h-80">
                  <p className="text-gray-500 dark:text-gray-400">No campaign data available</p>
                </div>
              )}

              {/* Performance by Placement */}
              {isLoadingPlacement ? (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 h-80 animate-pulse">
                  <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                  <div className="h-full bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              ) : placementReport?.reports && placementReport.reports.length > 0 ? (
                <BarChartComponent
                  data={placementReport.reports.map((r) => ({
                    name: r.slot_id || 'Unknown',
                    value: r.impressions,
                  }))}
                  title="Impressions by Placement"
                  valueLabel="Impressions"
                  color="#8b5cf6"
                  height={350}
                />
              ) : (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 flex items-center justify-center h-80">
                  <p className="text-gray-500 dark:text-gray-400">No placement data available</p>
                </div>
              )}
            </div>

            {/* Campaign Performance Table */}
            <div className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
                Campaign Performance
              </h2>
              {isLoadingCTR ? (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse">
                  <div className="h-40 bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              ) : ctrReport?.reports && ctrReport.reports.length > 0 ? (
                <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                      <thead className="bg-gray-50 dark:bg-gray-900">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Campaign
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Impressions
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Viewable
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Clicks
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            CTR
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Viewability
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Spend
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                        {ctrReport.reports.map((report) => (
                          <tr key={report.ad_id}>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                              {report.ad_name}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.impressions.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.viewable_impressions.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.clicks.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatPercentage(report.ctr)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatPercentage(report.viewability_rate)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatCurrency(report.spend_cents)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                  <p className="text-gray-500 dark:text-gray-400 text-center">
                    No campaign performance data available yet.
                  </p>
                </div>
              )}
            </div>

            {/* Placement Performance Table */}
            <div className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
                Placement Performance
              </h2>
              {isLoadingPlacement ? (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse">
                  <div className="h-40 bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              ) : placementReport?.reports && placementReport.reports.length > 0 ? (
                <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                      <thead className="bg-gray-50 dark:bg-gray-900">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Placement
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Impressions
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Clicks
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            CTR
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Viewability
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                            Unique Ads
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                        {placementReport.reports.map((report) => (
                          <tr key={report.slot_id}>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                              {report.slot_id || 'Unassigned'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.impressions.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.clicks.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatPercentage(report.ctr)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatPercentage(report.viewability_rate)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {report.unique_ads}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                  <p className="text-gray-500 dark:text-gray-400 text-center">
                    No placement performance data available yet.
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Campaign Modal */}
        {(modalType === 'create' || modalType === 'edit') && (
          <CampaignModal
            campaign={selectedCampaign}
            isOpen={true}
            onClose={handleCloseModal}
            onSubmit={(data) => {
              if (modalType === 'create') {
                createMutation.mutate(data as CreateCampaignRequest);
              } else if (selectedCampaign) {
                updateMutation.mutate({ id: selectedCampaign.id, data: data as UpdateCampaignRequest });
              }
            }}
            isLoading={createMutation.isPending || updateMutation.isPending}
            error={createMutation.error?.message || updateMutation.error?.message}
          />
        )}

        {/* Delete Confirmation Modal */}
        {modalType === 'delete' && selectedCampaign && (
          <DeleteConfirmationModal
            campaign={selectedCampaign}
            isOpen={true}
            onClose={handleCloseModal}
            onConfirm={() => deleteMutation.mutate(selectedCampaign.id)}
            isLoading={deleteMutation.isPending}
          />
        )}
      </div>
    </>
  );
};

// Campaign Modal Component
interface CampaignModalProps {
  campaign: Campaign | null;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateCampaignRequest | UpdateCampaignRequest) => void;
  isLoading: boolean;
  error?: string;
}

const CampaignModal: React.FC<CampaignModalProps> = ({
  campaign,
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  error,
}) => {
  const [formData, setFormData] = useState<CreateCampaignRequest>({
    name: campaign?.name || '',
    advertiser_name: campaign?.advertiser_name || '',
    ad_type: campaign?.ad_type || 'banner',
    content_url: campaign?.content_url || '',
    click_url: campaign?.click_url || '',
    alt_text: campaign?.alt_text || '',
    width: campaign?.width,
    height: campaign?.height,
    priority: campaign?.priority || 1,
    weight: campaign?.weight || 100,
    daily_budget_cents: campaign?.daily_budget_cents,
    total_budget_cents: campaign?.total_budget_cents,
    cpm_cents: campaign?.cpm_cents || 100,
    is_active: campaign?.is_active ?? true,
    start_date: campaign?.start_date?.split('T')[0],
    end_date: campaign?.end_date?.split('T')[0],
  });

  // Reset form state when campaign prop changes (for editing different campaigns)
  useEffect(() => {
    queueMicrotask(() => {
      setFormData({
        name: campaign?.name || '',
        advertiser_name: campaign?.advertiser_name || '',
        ad_type: campaign?.ad_type || 'banner',
        content_url: campaign?.content_url || '',
        click_url: campaign?.click_url || '',
        alt_text: campaign?.alt_text || '',
        width: campaign?.width,
        height: campaign?.height,
        priority: campaign?.priority || 1,
        weight: campaign?.weight || 100,
        daily_budget_cents: campaign?.daily_budget_cents,
        total_budget_cents: campaign?.total_budget_cents,
        cpm_cents: campaign?.cpm_cents || 100,
        is_active: campaign?.is_active ?? true,
        start_date: campaign?.start_date?.split('T')[0],
        end_date: campaign?.end_date?.split('T')[0],
      });
    });
  }, [campaign]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  const handleBannerSizeSelect = (width: number, height: number) => {
    setFormData({ ...formData, width, height });
  };

  if (!isOpen) return null;

  return (
    <Modal open={isOpen} onClose={onClose} title={campaign ? 'Edit Campaign' : 'Create Campaign'} size="xl">

          {error && (
            <div className="mb-4 bg-red-100 dark:bg-red-900 border border-red-400 text-red-700 dark:text-red-200 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Campaign Name *
                </label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Advertiser Name *
                </label>
                <input
                  type="text"
                  required
                  value={formData.advertiser_name}
                  onChange={(e) => setFormData({ ...formData, advertiser_name: e.target.value })}
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Ad Type *
                </label>
                <select
                  value={formData.ad_type}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      ad_type: e.target.value as 'banner' | 'video' | 'native',
                    })
                  }
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                >
                  <option value="banner">Banner</option>
                  <option value="video">Video</option>
                  <option value="native">Native</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  CPM (cents) *
                </label>
                <input
                  type="number"
                  min="1"
                  value={formData.cpm_cents}
                  onChange={(e) =>
                    setFormData({ ...formData, cpm_cents: parseInt(e.target.value) || 100 })
                  }
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Content URL *
              </label>
              <input
                type="url"
                required
                value={formData.content_url}
                onChange={(e) => setFormData({ ...formData, content_url: e.target.value })}
                placeholder="https://example.com/ad-image.jpg"
                className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Click URL
              </label>
              <input
                type="url"
                value={formData.click_url || ''}
                onChange={(e) => setFormData({ ...formData, click_url: e.target.value })}
                placeholder="https://example.com/landing-page"
                className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
              />
            </div>

            {formData.ad_type === 'banner' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Banner Size
                </label>
                <div className="flex flex-wrap gap-2 mb-2">
                  {STANDARD_BANNER_SIZES.slice(0, 6).map((size) => (
                    <button
                      key={`${size.width}x${size.height}`}
                      type="button"
                      onClick={() => handleBannerSizeSelect(size.width, size.height)}
                      className={`px-2 py-1 text-xs rounded ${
                        formData.width === size.width && formData.height === size.height
                          ? 'bg-primary text-white'
                          : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                      }`}
                    >
                      {size.width}x{size.height}
                    </button>
                  ))}
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <input
                      type="number"
                      min="1"
                      placeholder="Width"
                      value={formData.width || ''}
                      onChange={(e) =>
                        setFormData({ ...formData, width: parseInt(e.target.value) || undefined })
                      }
                      className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <input
                      type="number"
                      min="1"
                      placeholder="Height"
                      value={formData.height || ''}
                      onChange={(e) =>
                        setFormData({ ...formData, height: parseInt(e.target.value) || undefined })
                      }
                      className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                    />
                  </div>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Daily Budget (cents)
                </label>
                <input
                  type="number"
                  min="0"
                  value={formData.daily_budget_cents || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      daily_budget_cents: e.target.value ? parseInt(e.target.value) : undefined,
                    })
                  }
                  placeholder="Leave empty for unlimited"
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Total Budget (cents)
                </label>
                <input
                  type="number"
                  min="0"
                  value={formData.total_budget_cents || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      total_budget_cents: e.target.value ? parseInt(e.target.value) : undefined,
                    })
                  }
                  placeholder="Leave empty for unlimited"
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Start Date
                </label>
                <input
                  type="date"
                  value={formData.start_date || ''}
                  onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  End Date
                </label>
                <input
                  type="date"
                  value={formData.end_date || ''}
                  onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                  className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="flex items-center">
              <input
                type="checkbox"
                id="is_active"
                checked={formData.is_active}
                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
              />
              <label
                htmlFor="is_active"
                className="ml-2 block text-sm text-gray-700 dark:text-gray-300"
              >
                Active
              </label>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90 disabled:opacity-50"
              >
                {isLoading ? 'Saving...' : campaign ? 'Update Campaign' : 'Create Campaign'}
              </button>
            </div>
          </form>
    </Modal>
  );
};

// Delete Confirmation Modal
interface DeleteConfirmationModalProps {
  campaign: Campaign;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
}

const DeleteConfirmationModal: React.FC<DeleteConfirmationModalProps> = ({
  campaign,
  isOpen,
  onClose,
  onConfirm,
  isLoading,
}) => {
  if (!isOpen) return null;

  return (
    <Modal open={isOpen} onClose={onClose} title="Delete Campaign" size="md">
          <p
            className="text-sm text-gray-500 dark:text-gray-400 mb-4"
          >
            Are you sure you want to delete "{campaign.name}"? This action cannot be undone.
          </p>

          <div className="flex justify-end space-x-3">
            <Button
              onClick={onClose}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              onClick={onConfirm}
              disabled={isLoading}
              variant="danger"
            >
              {isLoading ? 'Deleting...' : 'Delete'}
            </Button>
          </div>
    </Modal>
  );
};

export default AdminCampaignsPage;
