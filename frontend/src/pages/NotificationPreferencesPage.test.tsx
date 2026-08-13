import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ToggleSwitch } from './NotificationPreferencesPage';

describe('notification preference switch', () => {
    it('exposes its visible label and checked state to assistive technology', () => {
        render(
            <ToggleSwitch
                label='New device login'
                description='Notify me after a new device signs in.'
                checked
                onChange={vi.fn()}
            />,
        );

        expect(
            screen.getByRole('switch', { name: 'New device login' }),
        ).toHaveAttribute('aria-checked', 'true');
    });
});
