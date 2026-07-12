import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { UserRoleBadge } from './UserRoleBadge';

describe('UserRoleBadge', () => {
  describe('Role variants', () => {
    it('renders admin badge with error variant (red)', () => {
      render(<UserRoleBadge role="admin" size="sm" />);
      const badge = screen.getByText('Admin');
      
      // Admin should use error variant (red)
      expect(badge.className).toContain('bg-error-100');
      expect(badge.className).toContain('text-error-800');
    });

    it('renders moderator badge with warning variant (yellow)', () => {
      render(<UserRoleBadge role="moderator" size="sm" />);
      const badge = screen.getByText('Moderator');
      
      // Moderator should use warning variant (yellow)
      expect(badge.className).toContain('bg-warning-100');
      expect(badge.className).toContain('text-warning-800');
    });

    it('renders creator badge with primary variant (blue)', () => {
      render(<UserRoleBadge role="creator" size="sm" />);
      const badge = screen.getByText('Creator');
      
      // Creator should use primary variant (blue)
      expect(badge.className).toContain('bg-primary-100');
      expect(badge.className).toContain('text-primary-800');
    });

    it('renders user badge with default variant (gray)', () => {
      render(<UserRoleBadge role="user" size="sm" />);
      const badge = screen.getByText('User');
      
      // User should use default variant (gray)
      expect(badge.className).toContain('bg-neutral-100');
      expect(badge.className).toContain('text-neutral-800');
    });

    it('renders member badge with default variant (gray)', () => {
      render(<UserRoleBadge role="member" size="sm" />);
      const badge = screen.getByText('Member');
      
      // Member should use default variant (gray)
      expect(badge.className).toContain('bg-neutral-100');
      expect(badge.className).toContain('text-neutral-800');
    });
  });

  describe('Size variants', () => {
    it('renders small badge correctly', () => {
      render(<UserRoleBadge role="admin" size="sm" />);
      const badge = screen.getByText('Admin');
      
      expect(badge.className).toContain('px-1.5');
      expect(badge.className).toContain('py-0.5');
      expect(badge.className).toContain('text-[11px]');
    });

    it('renders medium badge correctly', () => {
      render(<UserRoleBadge role="admin" size="md" />);
      const badge = screen.getByText('Admin');
      
      expect(badge.className).toContain('px-2');
      expect(badge.className).toContain('py-0.5');
      expect(badge.className).toContain('text-xs');
    });

    it('renders large badge correctly', () => {
      render(<UserRoleBadge role="admin" size="lg" />);
      const badge = screen.getByText('Admin');
      
      expect(badge.className).toContain('px-2.5');
      expect(badge.className).toContain('py-1');
      expect(badge.className).toContain('text-sm');
    });
  });

  describe('Tooltips', () => {
    it('shows admin tooltip by default', () => {
      render(<UserRoleBadge role="admin" size="sm" />);
      const badge = screen.getByText('Admin');
      
      expect(badge).toHaveAttribute(
        'title',
        'Administrator - Full access to all platform features and moderation tools'
      );
    });

    it('shows moderator tooltip by default', () => {
      render(<UserRoleBadge role="moderator" size="sm" />);
      const badge = screen.getByText('Moderator');
      
      expect(badge).toHaveAttribute(
        'title',
        'Moderator - Can review submissions, manage content, and enforce community guidelines'
      );
    });

    it('shows creator tooltip by default', () => {
      render(<UserRoleBadge role="creator" size="sm" />);
      const badge = screen.getByText('Creator');
      
      expect(badge).toHaveAttribute(
        'title',
        'Content Creator - Created or submitted this clip'
      );
    });

    it('shows user tooltip by default', () => {
      render(<UserRoleBadge role="user" size="sm" />);
      const badge = screen.getByText('User');
      
      expect(badge).toHaveAttribute(
        'title',
        'User - Regular community member'
      );
    });

    it('hides tooltip when showTooltip is false', () => {
      render(<UserRoleBadge role="admin" size="sm" showTooltip={false} />);
      const badge = screen.getByText('Admin');
      
      expect(badge).not.toHaveAttribute('title');
    });
  });

  describe('Accessibility', () => {
    it('renders as a span element', () => {
      const { container } = render(<UserRoleBadge role="admin" size="sm" />);
      const badge = container.querySelector('span');
      
      expect(badge).toBeInTheDocument();
    });

    it('applies custom className when provided', () => {
      render(<UserRoleBadge role="admin" size="sm" className="custom-class" />);
      const badge = screen.getByText('Admin');
      
      expect(badge.className).toContain('custom-class');
    });
  });
});
