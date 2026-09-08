import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';

describe('Button behavior', () => {
    it('activates with the keyboard without submitting its containing form', async () => {
        const user = userEvent.setup();
        const activate = vi.fn();
        const submit = vi.fn(event => event.preventDefault());
        render(<form onSubmit={submit}><Button onClick={activate}>Save</Button></form>);
        await user.tab();
        expect(screen.getByRole('button', { name: 'Save' })).toHaveFocus();
        await user.keyboard('{Enter} ');
        expect(activate).toHaveBeenCalledTimes(2);
        expect(submit).not.toHaveBeenCalled();
    });

    it('submits when explicitly requested', async () => {
        const submit = vi.fn(event => event.preventDefault());
        render(<form onSubmit={submit}><Button type='submit'>Save</Button></form>);
        await userEvent.click(screen.getByRole('button', { name: 'Save' }));
        expect(submit).toHaveBeenCalledOnce();
    });

    it.each(['disabled', 'loading'] as const)('blocks activation while %s', async state => {
        const activate = vi.fn();
        render(<Button {...{ [state]: true }} onClick={activate}>Save</Button>);
        const button = screen.getByRole('button', { name: 'Save' });
        expect(button).toBeDisabled();
        if (state === 'loading') expect(button).toHaveAttribute('aria-busy', 'true');
        await userEvent.click(button);
        expect(activate).not.toHaveBeenCalled();
    });

    it('preserves link semantics without nesting interactive controls', () => {
        render(<Button asChild><a href='/destination'>Continue</a></Button>);
        expect(screen.getByRole('link', { name: 'Continue' })).toHaveAttribute('href', '/destination');
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
});
