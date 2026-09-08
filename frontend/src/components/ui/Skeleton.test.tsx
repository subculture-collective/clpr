import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Skeleton } from './Skeleton';

describe('loading placeholders', () => {
    it.each(['text', 'circular', 'rectangular'] as const)('%s announces loading politely', variant => {
        const { container } = render(<Skeleton variant={variant} />);
        expect(container.firstElementChild).toHaveAttribute('aria-busy', 'true');
        expect(container.firstElementChild).toHaveAttribute('aria-live', 'polite');
    });
});
