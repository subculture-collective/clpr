import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  describe('Variant classes', () => {
    it('applies primary variant classes for selected state', () => {
      render(<Button variant="primary">Primary Button</Button>);
      const button = screen.getByRole('button', { name: 'Primary Button' });
      
      // Check for key classes that ensure visibility in both themes
      expect(button.className).toContain('bg-primary-500');
      expect(button.className).toContain('text-white');
      expect(button.className).toContain('dark:bg-primary-600');
      expect(button.className).toContain('dark:text-white');
      expect(button.className).toContain('shadow-sm');
    });

    it('applies ghost variant classes with clear hover states', () => {
      render(<Button variant="ghost">Ghost Button</Button>);
      const button = screen.getByRole('button', { name: 'Ghost Button' });
      
      // Check for hover and active states
      expect(button.className).toContain('bg-transparent');
      expect(button.className).toContain('hover:bg-neutral-100');
      expect(button.className).toContain('active:bg-neutral-200');
      expect(button.className).toContain('dark:hover:bg-neutral-800');
      expect(button.className).toContain('dark:active:bg-neutral-700');
      expect(button.className).toContain('text-foreground');
    });

    it('applies focus-visible ring with proper offset for light theme', () => {
      render(<Button>Focus Test</Button>);
      const button = screen.getByRole('button', { name: 'Focus Test' });
      
      // Check focus-visible classes for WCAG AA compliance
      expect(button.className).toContain('focus-visible:ring-2');
      expect(button.className).toContain('focus-visible:ring-primary-500');
      expect(button.className).toContain('focus-visible:ring-offset-2');
    });

    it('applies focus-visible ring with dark mode offset', () => {
      render(<Button>Dark Focus Test</Button>);
      const button = screen.getByRole('button', { name: 'Dark Focus Test' });
      
      // Check dark mode focus-visible offset
      expect(button.className).toContain('dark:focus-visible:ring-offset-neutral-900');
    });

    it('applies secondary variant classes', () => {
      render(<Button variant="secondary">Secondary Button</Button>);
      const button = screen.getByRole('button', { name: 'Secondary Button' });
      
      expect(button.className).toContain('bg-secondary-500');
      expect(button.className).toContain('text-white');
    });

    it('applies outline variant classes', () => {
      render(<Button variant="outline">Outline Button</Button>);
      const button = screen.getByRole('button', { name: 'Outline Button' });
      
      expect(button.className).toContain('border-2');
      expect(button.className).toContain('border-primary-500');
      expect(button.className).toContain('text-primary-500');
      expect(button.className).toContain('hover:bg-primary-50');
      expect(button.className).toContain('dark:hover:bg-primary-950');
    });

    it('applies danger variant classes', () => {
      render(<Button variant="danger">Danger Button</Button>);
      const button = screen.getByRole('button', { name: 'Danger Button' });
      
      expect(button.className).toContain('bg-error-500');
      expect(button.className).toContain('text-white');
    });
  });

  describe('Size classes', () => {
    it('applies small size classes', () => {
      render(<Button size="sm">Small</Button>);
      const button = screen.getByRole('button', { name: 'Small' });
      
      expect(button.className).toContain('px-3');
      expect(button.className).toContain('py-2');
      expect(button.className).toContain('text-xs');
      expect(button.className).toContain('min-h-[44px]');
    });

    it('applies medium size classes by default', () => {
      render(<Button>Medium</Button>);
      const button = screen.getByRole('button', { name: 'Medium' });
      
      expect(button.className).toContain('px-4');
      expect(button.className).toContain('py-2.5');
      expect(button.className).toContain('text-sm');
      expect(button.className).toContain('min-h-[44px]');
    });

    it('applies large size classes', () => {
      render(<Button size="lg">Large</Button>);
      const button = screen.getByRole('button', { name: 'Large' });
      
      expect(button.className).toContain('px-6');
      expect(button.className).toContain('py-3');
      expect(button.className).toContain('text-base');
      expect(button.className).toContain('min-h-[48px]');
    });
  });

  describe('State classes', () => {
    it('does not submit a form unless explicitly requested', () => {
      render(<Button>Safe default</Button>);
      expect(screen.getByRole('button', { name: 'Safe default' })).toHaveAttribute('type', 'button');
    });

    it('applies disabled state classes', () => {
      render(<Button disabled>Disabled</Button>);
      const button = screen.getByRole('button', { name: 'Disabled' });
      
      expect(button).toBeDisabled();
      expect(button.className).toContain('disabled:opacity-50');
      expect(button.className).toContain('disabled:cursor-not-allowed');
    });

    it('applies loading state and shows spinner', () => {
      render(<Button loading>Loading</Button>);
      const button = screen.getByRole('button', { name: 'Loading' });
      
      expect(button).toBeDisabled();
      expect(button).toHaveAttribute('aria-busy', 'true');
      expect(button.querySelector('svg.animate-spin')).toBeInTheDocument();
      expect(button.querySelector('svg.animate-spin')).toHaveAttribute('aria-hidden', 'true');
    });
  });

  describe('Full width', () => {
    it('applies full width class when specified', () => {
      render(<Button fullWidth>Full Width</Button>);
      const button = screen.getByRole('button', { name: 'Full Width' });
      
      expect(button.className).toContain('w-full');
    });
  });
});
