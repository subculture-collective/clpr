import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ReputationBadge, ReputationProgressBar } from './ReputationBadge';

describe('ReputationBadge', () => {
  it('renders new member badge correctly', () => {
    render(<ReputationBadge badge="new" score={10} />);
    
    expect(screen.getByText('New Member')).toBeInTheDocument();
    expect(screen.getByText('10 rep')).toBeInTheDocument();
  });

  it('renders contributor badge correctly', () => {
    render(<ReputationBadge badge="contributor" score={100} />);
    
    expect(screen.getByText('Contributor')).toBeInTheDocument();
    expect(screen.getByText('100 rep')).toBeInTheDocument();
  });

  it('renders expert badge correctly', () => {
    render(<ReputationBadge badge="expert" score={300} />);
    
    expect(screen.getByText('Expert')).toBeInTheDocument();
    expect(screen.getByText('300 rep')).toBeInTheDocument();
  });

  it('renders moderator badge correctly', () => {
    render(<ReputationBadge badge="moderator" score={500} />);
    
    expect(screen.getByText('Moderator')).toBeInTheDocument();
    expect(screen.getByText('500 rep')).toBeInTheDocument();
  });

  it('hides score when showScore is false', () => {
    render(<ReputationBadge badge="expert" score={300} showScore={false} />);
    
    expect(screen.getByText('Expert')).toBeInTheDocument();
    expect(screen.queryByText('300 rep')).not.toBeInTheDocument();
  });

  it('applies correct colors for different badges', () => {
    const { rerender } = render(<ReputationBadge badge="new" />);
    expect(screen.getByText('New Member')).toHaveClass('bg-muted');

    rerender(<ReputationBadge badge="contributor" />);
    expect(screen.getByText('Contributor')).toHaveClass('bg-blue-500');

    rerender(<ReputationBadge badge="expert" />);
    expect(screen.getByText('Expert')).toHaveClass('bg-yellow-500');

    rerender(<ReputationBadge badge="moderator" />);
    expect(screen.getByText('Moderator')).toHaveClass('bg-red-500');
  });
});

describe('ReputationProgressBar', () => {
  it('renders progress bar with score', () => {
    render(<ReputationProgressBar score={25} />);
    
    expect(screen.getByText('25 reputation')).toBeInTheDocument();
    expect(screen.getByText(/25 to Contributor/)).toBeInTheDocument();
  });

  it('shows progress to contributor for new members', () => {
    render(<ReputationProgressBar score={10} />);
    
    expect(screen.getByText('10 reputation')).toBeInTheDocument();
    expect(screen.getByText('40 to Contributor')).toBeInTheDocument();
  });

  it('shows progress to expert for contributors', () => {
    render(<ReputationProgressBar score={100} />);
    
    expect(screen.getByText('100 reputation')).toBeInTheDocument();
    expect(screen.getByText('150 to Expert')).toBeInTheDocument();
  });

  it('shows max level for experts', () => {
    render(<ReputationProgressBar score={300} />);
    
    expect(screen.getByText('300 reputation')).toBeInTheDocument();
    expect(screen.queryByText(/to/)).not.toBeInTheDocument();
  });

  it('calculates progress percentage correctly', () => {
    const { container } = render(<ReputationProgressBar score={25} />);
    
    // 25 out of 50 = 50% progress
    const progressBar = container.querySelector('[style*="width"]');
    expect(progressBar).toHaveStyle({ width: '50%' });
  });

  it('applies correct color for new member tier', () => {
    const { container } = render(<ReputationProgressBar score={25} />);
    const progressBar = container.querySelector('[style*="width"]');
    expect(progressBar).toHaveClass('bg-blue-500');
  });

  it('applies correct color for contributor tier', () => {
    const { container } = render(<ReputationProgressBar score={100} />);
    const progressBar = container.querySelector('[style*="width"]');
    expect(progressBar).toHaveClass('bg-yellow-500');
  });

  it('applies correct color for expert tier', () => {
    const { container } = render(<ReputationProgressBar score={300} />);
    const progressBar = container.querySelector('[style*="width"]');
    expect(progressBar).toHaveClass('bg-green-500');
  });
});
