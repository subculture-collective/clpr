import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Avatar } from './Avatar';

describe('Avatar', () => {
    it('replaces a broken image with an initial instead of alt text', () => {
        const { container } = render(<Avatar src='https://example.com/broken.png' alt='moonmoon' />);

        fireEvent.error(container.querySelector('img')!);

        expect(container.querySelector('img')).toBeNull();
        expect(screen.getByTestId('avatar-fallback')).toHaveTextContent('M');
        expect(screen.getByRole('img', { name: 'moonmoon' })).toBeInTheDocument();
    });

    it('prefers the fallback name and keeps decorative avatars silent', () => {
        render(<Avatar alt='' fallback='@dani_denos' />);

        expect(screen.getByTestId('avatar-fallback')).toHaveTextContent('D');
        expect(screen.queryByRole('img')).toBeNull();
    });

    it('tries a new image after the previous source failed', () => {
        const { container, rerender } = render(<Avatar src='https://example.com/a.png' alt='a' />);
        fireEvent.error(container.querySelector('img')!);
        expect(container.querySelector('img')).toBeNull();

        rerender(<Avatar src='https://example.com/b.png' alt='a' />);

        expect(container.querySelector('img')).toHaveAttribute('src', 'https://example.com/b.png');
    });

    it('applies custom frame sizes over the preset size', () => {
        render(<Avatar alt='' fallback='x' frameClassName='h-24 w-24' />);

        const frame = screen.getByTestId('avatar-fallback').parentElement!;
        expect(frame.className).toContain('h-24');
        expect(frame.className).not.toContain('h-10');
    });
});
