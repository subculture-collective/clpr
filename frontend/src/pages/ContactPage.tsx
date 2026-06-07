import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Container, Card, CardBody, Button, Input, TextArea, Alert } from '../components';
import { useAuth } from '../context/AuthContext';
import { submitContactMessage, type ContactMessageRequest } from '../lib/contact-api';

export function ContactPage() {
  const { user, isAuthenticated } = useAuth();
  const [formData, setFormData] = useState<ContactMessageRequest>({
    email: user?.email || '',
    category: 'feedback',
    subject: '',
    message: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitSuccess, setSubmitSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await submitContactMessage(formData);
      setSubmitSuccess(true);
      // Reset form
      setFormData({
        email: user?.email || '',
        category: 'feedback',
        subject: '',
        message: '',
      });
    } catch (err) {
      if (err && typeof err === 'object' && 'response' in err) {
        const response = (err as { response?: { data?: { error?: string } } }).response;
        setError(response?.data?.error || 'Failed to submit contact message. Please try again.');
      } else {
        setError('Failed to submit contact message. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleChange = (field: keyof ContactMessageRequest, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear success message when user starts editing again
    if (submitSuccess) {
      setSubmitSuccess(false);
    }
  };

  return (
    <Container className="py-8 max-w-3xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">Contact Us</h1>
        <p className="text-muted-foreground">
          Have a question, feedback, or need help? We're here to assist you.
        </p>
      </div>

      {submitSuccess && (
        <Alert variant="success" className="mb-6">
          <div>
            <p className="font-semibold">Message sent successfully!</p>
            <p className="text-sm mt-1">
              We've received your message and will get back to you as soon as possible.
            </p>
          </div>
        </Alert>
      )}

      {error && (
        <Alert variant="error" className="mb-6">
          {error}
        </Alert>
      )}

      <Card>
        <CardBody>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Category Selection */}
            <div className="flex flex-col gap-1.5">
              <label htmlFor="category" className="text-sm font-medium text-foreground">
                Category *
              </label>
              <select
                id="category"
                value={formData.category}
                onChange={(e) => handleChange('category', e.target.value)}
                className="w-full px-3 py-2.5 rounded-lg border transition-colors min-h-[44px] bg-background text-foreground border-border hover:border-primary-300 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                required
              >
                <option value="feedback">General Feedback</option>
                <option value="account">Account Help</option>
                <option value="billing">Billing Question</option>
                <option value="abuse">Report Abuse</option>
              </select>
              <p className="text-sm text-muted-foreground">
                Select the category that best describes your inquiry
              </p>
            </div>

            {/* Email Input */}
            <Input
              label="Email Address *"
              type="email"
              id="email"
              value={formData.email}
              onChange={(e) => handleChange('email', e.target.value)}
              placeholder="your.email@example.com"
              required
              fullWidth
              helperText={
                isAuthenticated
                  ? "We'll use this email to respond to your message"
                  : 'Enter your email so we can respond to you'
              }
            />

            {/* Subject Input */}
            <Input
              label="Subject *"
              type="text"
              id="subject"
              value={formData.subject}
              onChange={(e) => handleChange('subject', e.target.value)}
              placeholder="Brief description of your inquiry"
              required
              fullWidth
              maxLength={200}
              helperText="A brief summary of your message (3-200 characters)"
            />

            {/* Message TextArea */}
            <TextArea
              label="Message *"
              id="message"
              value={formData.message}
              onChange={(e) => handleChange('message', e.target.value)}
              placeholder="Please provide details about your inquiry..."
              required
              fullWidth
              maxLength={5000}
              showCount
              rows={8}
              helperText="Please provide as much detail as possible (10-5000 characters)"
            />

            {/* Privacy Notice */}
            <div className="bg-accent/50 border border-border rounded-lg p-4">
              <p className="text-sm text-muted-foreground">
                <strong className="text-foreground">Privacy Notice:</strong> By submitting this form, 
                you agree to our{' '}
                <Link to="/privacy" className="text-primary hover:underline">
                  Privacy Policy
                </Link>
                . We will only use your contact information to respond to your inquiry and will not 
                share it with third parties without your consent.
              </p>
            </div>

            {/* Submit Button */}
            <div className="flex gap-4">
              <Button
                type="submit"
                disabled={isSubmitting}
                className="min-w-[120px]"
              >
                {isSubmitting ? 'Sending...' : 'Send Message'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setFormData({
                    email: user?.email || '',
                    category: 'feedback',
                    subject: '',
                    message: '',
                  });
                  setError(null);
                  setSubmitSuccess(false);
                }}
                disabled={isSubmitting}
              >
                Clear Form
              </Button>
            </div>
          </form>
        </CardBody>
      </Card>

      {/* Additional Help Section */}
      <Card className="mt-6">
        <CardBody>
          <h2 className="text-xl font-semibold mb-4">Other Ways to Get Help</h2>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium text-foreground mb-1">Community Guidelines</h3>
              <p className="text-sm text-muted-foreground">
                Review our{' '}
                <Link to="/community-rules" className="text-primary hover:underline">
                  Community Rules
                </Link>{' '}
                for information on acceptable behavior and content standards.
              </p>
            </div>
            <div>
              <h3 className="font-medium text-foreground mb-1">Report Content</h3>
              <p className="text-sm text-muted-foreground">
                To report specific clips or comments, use the report button on the content itself 
                for faster moderation.
              </p>
            </div>
            <div>
              <h3 className="font-medium text-foreground mb-1">GitHub Issues</h3>
              <p className="text-sm text-muted-foreground">
                For bug reports or feature requests, you can also open an issue on our{' '}
                <a
                  href="https://git.subcult.tv/subculture-collective/clpr/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  GitHub repository
                </a>
                .
              </p>
            </div>
          </div>
        </CardBody>
      </Card>
    </Container>
  );
}
