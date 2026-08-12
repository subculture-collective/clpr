import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { ContactPage } from './ContactPage';
import * as contactApi from '../lib/contact-api';

// Mock the contact API
vi.mock('../lib/contact-api');

// Mock AuthContext
vi.mock('../context/AuthContext', async () => {
    const actual = await vi.importActual('../context/AuthContext');
    return {
        ...actual,
        useAuth: () => ({
            user: null,
            isAuthenticated: false,
        }),
    };
});

const renderContactPage = () => {
    return render(
        <BrowserRouter>
            <ContactPage />
        </BrowserRouter>
    );
};

describe('ContactPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders contact form with all fields', () => {
        renderContactPage();

        expect(screen.getByText('Contact Us')).toBeInTheDocument();
        expect(screen.getByLabelText(/Category/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/Email Address/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/Subject/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/Message/i)).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: /Send Message/i })
        ).toBeInTheDocument();
    });

    it('displays validation errors for empty required fields', async () => {
        const user = userEvent.setup();
        renderContactPage();

        const submitButton = screen.getByRole('button', {
            name: /Send Message/i,
        });
        await user.click(submitButton);

        // Form should show browser validation since fields are required
        const emailInput = screen.getByLabelText(
            /Email Address/i
        ) as HTMLInputElement;
        expect(emailInput.validity.valid).toBe(false);
    });

    it('allows user to fill out the form', async () => {
        const user = userEvent.setup();
        renderContactPage();

        // Fill out the form
        const categorySelect = screen.getByLabelText(/Category/i);
        await user.selectOptions(categorySelect, 'feedback');

        const emailInput = screen.getByLabelText(/Email Address/i);
        await user.type(emailInput, 'test@example.com');

        const subjectInput = screen.getByLabelText(/Subject/i);
        await user.type(subjectInput, 'Test Subject');

        const messageTextarea = screen.getByLabelText(/Message/i);
        await user.type(
            messageTextarea,
            'This is a test message for the contact form.'
        );

        expect(emailInput).toHaveValue('test@example.com');
        expect(subjectInput).toHaveValue('Test Subject');
        expect(messageTextarea).toHaveValue(
            'This is a test message for the contact form.'
        );
    });

    it('submits form and shows success message', async () => {
        const user = userEvent.setup();
        const mockSubmit = vi.mocked(contactApi.submitContactMessage);
        mockSubmit.mockResolvedValueOnce({
            message: 'Message sent successfully',
            status: 'success',
        });

        renderContactPage();

        // Fill out the form
        await user.selectOptions(
            screen.getByLabelText(/Category/i),
            'feedback'
        );
        await user.type(
            screen.getByLabelText(/Email Address/i),
            'test@example.com'
        );
        await user.type(screen.getByLabelText(/Subject/i), 'Test Subject');
        await user.type(
            screen.getByLabelText(/Message/i),
            'This is a test message for the contact form.'
        );

        // Submit the form
        const submitButton = screen.getByRole('button', {
            name: /Send Message/i,
        });
        await user.click(submitButton);

        // Wait for success message
        await waitFor(() => {
            expect(
                screen.getByText(/Message sent successfully/i)
            ).toBeInTheDocument();
        });

        // Verify API was called
        expect(mockSubmit).toHaveBeenCalledWith({
            email: 'test@example.com',
            category: 'feedback',
            subject: 'Test Subject',
            message: 'This is a test message for the contact form.',
        });
    });

    it('shows error message on submission failure', async () => {
        const user = userEvent.setup();
        const mockSubmit = vi.mocked(contactApi.submitContactMessage);
        mockSubmit.mockRejectedValueOnce({
            response: { data: { error: 'Server error' } },
        });

        renderContactPage();

        // Fill out the form
        await user.selectOptions(
            screen.getByLabelText(/Category/i),
            'feedback'
        );
        await user.type(
            screen.getByLabelText(/Email Address/i),
            'test@example.com'
        );
        await user.type(screen.getByLabelText(/Subject/i), 'Test Subject');
        await user.type(
            screen.getByLabelText(/Message/i),
            'This is a test message for the contact form.'
        );

        // Submit the form
        const submitButton = screen.getByRole('button', {
            name: /Send Message/i,
        });
        await user.click(submitButton);

        // Wait for error message
        await waitFor(() => {
            expect(screen.getByText(/Server error/i)).toBeInTheDocument();
        });
    });

    it('clears form when clear button is clicked', async () => {
        const user = userEvent.setup();
        renderContactPage();

        // Fill out the form
        const emailInput = screen.getByLabelText(
            /Email Address/i
        ) as HTMLInputElement;
        const subjectInput = screen.getByLabelText(
            /Subject/i
        ) as HTMLInputElement;
        const messageTextarea = screen.getByLabelText(
            /Message/i
        ) as HTMLTextAreaElement;

        await user.type(emailInput, 'test@example.com');
        await user.type(subjectInput, 'Test Subject');
        await user.type(messageTextarea, 'Test message');

        // Click clear button
        const clearButton = screen.getByRole('button', { name: /Clear Form/i });
        await user.click(clearButton);

        // Check fields are cleared
        expect(emailInput).toHaveValue('');
        expect(subjectInput).toHaveValue('');
        expect(messageTextarea).toHaveValue('');
    });

    it('displays privacy notice', () => {
        renderContactPage();

        expect(screen.getByText(/Privacy Notice:/i)).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: /Privacy Policy/i })
        ).toBeInTheDocument();
    });

    it('displays additional help section', () => {
        renderContactPage();

        expect(screen.getByText(/Other Ways to Get Help/i)).toBeInTheDocument();
        expect(screen.getByText(/Community Guidelines/i)).toBeInTheDocument();
        expect(screen.getByText(/Report Content/i)).toBeInTheDocument();
        expect(screen.getByText(/Email support/i)).toBeInTheDocument();
    });
});
