import { describe, it, expect } from 'vitest';
import { render, screen } from '@/test/test-utils';
import { ReputationDisplay, RankBadge } from './ReputationDisplay';
import type { UserReputation } from '@/types/reputation';

const mockReputation: UserReputation = {
  user_id: '123',
  username: 'testuser',
  display_name: 'Test User',
  avatar_url: 'https://example.com/avatar.jpg',
  karma_points: 1234,
  rank: 'Contributor',
  trust_score: 85,
  engagement_score: 5678,
  badges: [
    {
      id: '1',
      badge_id: 'veteran',
      awarded_at: '2024-01-01T00:00:00Z',
      name: 'Veteran',
      description: 'Member for over 1 year',
      icon: '🏆',
      category: 'achievement',
    },
  ],
  stats: {
    user_id: '123',
    trust_score: 85,
    engagement_score: 5678,
    total_comments: 100,
    total_votes_cast: 500,
    total_clips_submitted: 25,
    correct_reports: 10,
    incorrect_reports: 2,
    days_active: 150,
    updated_at: '2024-01-01T00:00:00Z',
  },
  created_at: '2023-01-01T00:00:00Z',
};

describe('ReputationDisplay', () => {
  it('displays user information', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByText('Contributor')).toBeInTheDocument();
  });

  it('displays karma points', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('1,234')).toBeInTheDocument();
    expect(screen.getByText('Total Karma')).toBeInTheDocument();
  });

  it('displays trust score', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('85')).toBeInTheDocument();
    expect(screen.getByText('Trust Score')).toBeInTheDocument();
  });

  it('displays engagement score', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('5,678')).toBeInTheDocument();
    expect(screen.getByText('Engagement')).toBeInTheDocument();
  });

  it('displays badges', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('🏆')).toBeInTheDocument();
    expect(screen.getByText('Badges')).toBeInTheDocument();
  });

  it('displays activity stats', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    expect(screen.getByText('100')).toBeInTheDocument(); // comments
    expect(screen.getByText('500')).toBeInTheDocument(); // votes
    expect(screen.getByText('25')).toBeInTheDocument(); // submissions
  });

  it('displays avatar when available', () => {
    render(<ReputationDisplay reputation={mockReputation} />);
    
    const avatar = screen.getByAltText('testuser');
    expect(avatar).toBeInTheDocument();
    expect(avatar).toHaveAttribute('src', 'https://example.com/avatar.jpg');
  });

  it('renders in compact mode', () => {
    render(<ReputationDisplay reputation={mockReputation} compact={true} />);
    
    // In compact mode, should show karma and rank
    expect(screen.getByText('1,234')).toBeInTheDocument();
    expect(screen.getByText('Contributor')).toBeInTheDocument();
    // But not full stats
    expect(screen.queryByText('Trust Score')).not.toBeInTheDocument();
  });

  it('shows karma label in compact mode', () => {
    render(<ReputationDisplay reputation={mockReputation} compact={true} />);
    
    expect(screen.getByText('karma')).toBeInTheDocument();
  });
});

describe('RankBadge', () => {
  it('renders rank name', () => {
    render(<RankBadge rank="Legend" />);
    
    expect(screen.getByText('Legend')).toBeInTheDocument();
  });

  it('applies correct color for different ranks', () => {
    const { rerender } = render(<RankBadge rank="Newcomer" />);
    expect(screen.getByText('Newcomer')).toHaveClass('text-muted-foreground');
    
    rerender(<RankBadge rank="Legend" />);
    expect(screen.getByText('Legend')).toHaveClass('text-red-400');
    
    rerender(<RankBadge rank="Veteran" />);
    expect(screen.getByText('Veteran')).toHaveClass('text-yellow-400');
  });

  it('renders in different sizes', () => {
    const { rerender } = render(<RankBadge rank="Member" size="sm" />);
    expect(screen.getByText('Member')).toHaveClass('text-xs');
    
    rerender(<RankBadge rank="Member" size="md" />);
    expect(screen.getByText('Member')).toHaveClass('text-sm');
    
    rerender(<RankBadge rank="Member" size="lg" />);
    expect(screen.getByText('Member')).toHaveClass('text-base');
  });
});
