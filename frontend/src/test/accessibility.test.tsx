import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { axe } from 'jest-axe';
import { Input } from '../components/ui/Input';
import { TextArea } from '../components/ui/TextArea';
import { Checkbox } from '../components/ui/Checkbox';
import { Toggle } from '../components/ui/Toggle';
import { Button } from '../components/ui/Button';
import { Modal } from '../components/ui/Modal';

// Focus trapping, Escape, focus restoration, and skip links have dedicated
// component tests. Layout, contrast, and touch targets belong in Playwright.
describe('accessible controls', () => {
    it.each([Input, TextArea])('%s associates labels and validation feedback', async Control => {
        const { container, rerender } = render(<Control label='Message' helperText='Tell us more' />);
        const field = screen.getByRole('textbox', { name: 'Message' });
        expect(field).toHaveAccessibleDescription('Tell us more');
        await userEvent.type(field, 'Hello');
        expect(field).toHaveValue('Hello');
        rerender(<Control label='Message' error='Message is too short' />);
        expect(field).toHaveAttribute('aria-invalid', 'true');
        expect(field).toHaveAccessibleDescription('Message is too short');
        expect((await axe(container)).violations).toEqual([]);
    });

    it.each([
        { Control: Checkbox, role: 'checkbox' },
        { Control: Toggle, role: 'switch' },
    ])('$role exposes state and supports keyboard changes', async ({ Control, role }) => {
        const change = vi.fn();
        const user = userEvent.setup();
        const { container, rerender } = render(<Control label='Notifications' checked={false} onChange={change} />);
        const control = screen.getByRole(role, { name: 'Notifications' });
        expect(control).not.toBeChecked();
        await user.tab();
        expect(control).toHaveFocus();
        await user.keyboard(' ');
        expect(change).toHaveBeenCalledOnce();
        rerender(<Control label='Notifications' checked onChange={change} />);
        expect(control).toBeChecked();
        expect((await axe(container)).violations).toEqual([]);
    });

    it('keeps form navigation in order and skips disabled controls', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <form aria-label='Contact'>
                <Input label='Name' />
                <Button disabled>Unavailable</Button>
                <TextArea label='Message' />
                <Button type='submit'>Send</Button>
            </form>,
        );
        await user.tab();
        expect(screen.getByRole('textbox', { name: 'Name' })).toHaveFocus();
        await user.tab();
        expect(screen.getByRole('textbox', { name: 'Message' })).toHaveFocus();
        await user.tab();
        expect(screen.getByRole('button', { name: 'Send' })).toHaveFocus();
        await user.tab({ shift: true });
        expect(screen.getByRole('textbox', { name: 'Message' })).toHaveFocus();
        expect((await axe(container)).violations).toEqual([]);
    });

    it('exposes a named modal dialog without axe violations', async () => {
        render(<Modal open title='Confirm action' onClose={() => {}}><Button>Confirm</Button></Modal>);
        const dialog = screen.getByRole('dialog', { name: 'Confirm action' });
        expect(dialog).toHaveAttribute('aria-modal', 'true');
        expect((await axe(dialog)).violations).toEqual([]);
    });
});
