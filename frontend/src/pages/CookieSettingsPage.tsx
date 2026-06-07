import { Helmet } from '@dr.pogodin/react-helmet';
import { Link } from 'react-router-dom';
import {
  Alert,
  Button,
  Card,
  CardBody,
  CardHeader,
  Container,
  Stack,
  Toggle,
} from '../components';
import { useConsent } from '../context/ConsentContext';
import { useState } from 'react';

/**
 * Cookie Settings Page
 * Dedicated page for managing cookie consent preferences
 * Displays detailed information about each cookie category
 */
export function CookieSettingsPage() {
  const { consent, updateConsent, doNotTrack, acceptAll, rejectAll } = useConsent();
  const [success, setSuccess] = useState(false);

  const handleConsentChange = (category: string, value: boolean) => {
    updateConsent({ [category]: value });
    setSuccess(true);
    setTimeout(() => setSuccess(false), 3000);
  };

  return (
    <>
      <Helmet>
        <title>Cookie Settings - clpr</title>
        <meta name="description" content="Manage your cookie and privacy preferences on clpr" />
      </Helmet>

      <Container className="py-4 xs:py-6 md:py-8">
        <div className="max-w-4xl mx-auto">
          <div className="mb-6">
            <h1 className="text-2xl xs:text-3xl font-bold mb-2">
              Cookie Settings
            </h1>
            <p className="text-muted-foreground">
              Manage your cookie and privacy preferences. You can choose which types of cookies we can use.
            </p>
          </div>

          {doNotTrack && (
            <Alert variant="info" className="mb-6">
              <strong>Do Not Track detected:</strong> Your browser has Do Not Track enabled. 
              Tracking cookies will be disabled regardless of your settings.
            </Alert>
          )}

          {success && (
            <Alert variant="success" className="mb-6">
              Your cookie preferences have been updated!
            </Alert>
          )}

          {/* Quick Actions */}
          <Card className="mb-6">
            <CardHeader>
              <h2 className="text-xl font-semibold">Quick Actions</h2>
            </CardHeader>
            <CardBody>
              <div className="flex flex-wrap gap-3">
                <Button
                  variant="primary"
                  onClick={acceptAll}
                >
                  Accept All Cookies
                </Button>
                <Button
                  variant="outline"
                  onClick={rejectAll}
                >
                  Reject All Optional Cookies
                </Button>
              </div>
            </CardBody>
          </Card>

          {/* Essential Cookies */}
          <Card className="mb-6">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">Essential Cookies</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Required for the website to function
                  </p>
                </div>
                <span className="text-xs bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300 px-3 py-1 rounded">
                  Always Active
                </span>
              </div>
            </CardHeader>
            <CardBody>
              <p className="text-sm text-muted-foreground mb-4">
                These cookies are necessary for the website to function and cannot be switched off. 
                They are usually only set in response to actions made by you such as setting your privacy preferences, 
                logging in, or filling in forms.
              </p>
              <table className="w-full text-sm">
                <thead className="border-b border-border">
                  <tr>
                    <th className="text-left py-2 font-medium">Cookie</th>
                    <th className="text-left py-2 font-medium">Purpose</th>
                    <th className="text-left py-2 font-medium">Duration</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">session_token</td>
                    <td className="py-2">Authentication and session management</td>
                    <td className="py-2">7 days</td>
                  </tr>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">csrf_token</td>
                    <td className="py-2">Security protection against cross-site attacks</td>
                    <td className="py-2">Session</td>
                  </tr>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">clpr_consent_preferences</td>
                    <td className="py-2">Stores your cookie consent choices</td>
                    <td className="py-2">12 months</td>
                  </tr>
                </tbody>
              </table>
            </CardBody>
          </Card>

          {/* Functional Cookies */}
          <Card className="mb-6">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">Functional Cookies</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Enhance your experience with personalized features
                  </p>
                </div>
                <Toggle
                  checked={consent.functional}
                  onChange={(e) => handleConsentChange('functional', e.target.checked)}
                  disabled={doNotTrack}
                />
              </div>
            </CardHeader>
            <CardBody>
              <p className="text-sm text-muted-foreground mb-4">
                These cookies allow us to remember choices you make (such as your language preference, 
                theme, or region) and provide enhanced, more personal features.
              </p>
              <table className="w-full text-sm">
                <thead className="border-b border-border">
                  <tr>
                    <th className="text-left py-2 font-medium">Cookie</th>
                    <th className="text-left py-2 font-medium">Purpose</th>
                    <th className="text-left py-2 font-medium">Duration</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">theme_preference</td>
                    <td className="py-2">Remembers your dark mode preference</td>
                    <td className="py-2">12 months</td>
                  </tr>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">language</td>
                    <td className="py-2">Remembers your language preference</td>
                    <td className="py-2">12 months</td>
                  </tr>
                </tbody>
              </table>
            </CardBody>
          </Card>

          {/* Analytics Cookies */}
          <Card className="mb-6">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">Analytics Cookies</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Help us understand how you use our site
                  </p>
                </div>
                <Toggle
                  checked={consent.analytics}
                  onChange={(e) => handleConsentChange('analytics', e.target.checked)}
                  disabled={doNotTrack}
                />
              </div>
            </CardHeader>
            <CardBody>
              <p className="text-sm text-muted-foreground mb-4">
                These cookies help us understand how visitors interact with our website by collecting 
                and reporting information anonymously. This helps us improve the site.
              </p>
              <table className="w-full text-sm">
                <thead className="border-b border-border">
                  <tr>
                    <th className="text-left py-2 font-medium">Cookie</th>
                    <th className="text-left py-2 font-medium">Purpose</th>
                    <th className="text-left py-2 font-medium">Duration</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">_ph_*</td>
                    <td className="py-2">PostHog analytics (first-party)</td>
                    <td className="py-2">12 months</td>
                  </tr>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">_ga</td>
                    <td className="py-2">Google Analytics (third-party)</td>
                    <td className="py-2">24 months</td>
                  </tr>
                  <tr className="border-b border-border">
                    <td className="py-2 font-mono text-xs">_ga_*</td>
                    <td className="py-2">Google Analytics (third-party)</td>
                    <td className="py-2">24 months</td>
                  </tr>
                </tbody>
              </table>
            </CardBody>
          </Card>

          {/* Advertising Cookies */}
          <Card className="mb-6">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">Advertising Cookies</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Personalize ads based on your interests
                  </p>
                </div>
                <Toggle
                  checked={consent.advertising}
                  onChange={(e) => handleConsentChange('advertising', e.target.checked)}
                  disabled={doNotTrack}
                />
              </div>
            </CardHeader>
            <CardBody>
              <p className="text-sm text-muted-foreground mb-4">
                These cookies are used to make advertising messages more relevant to you. 
                They perform functions like preventing the same ad from continuously reappearing, 
                ensuring that ads are properly displayed, and in some cases selecting ads based on your interests.
              </p>
              <p className="text-sm text-muted-foreground">
                Without these cookies, you will still see ads, but they will be less relevant to your interests.
              </p>
            </CardBody>
          </Card>

          {/* Additional Information */}
          <Card>
            <CardHeader>
              <h2 className="text-xl font-semibold">More Information</h2>
            </CardHeader>
            <CardBody>
              <Stack direction="vertical" gap={3}>
                <p className="text-sm text-muted-foreground">
                  For more information about how we use cookies and protect your privacy, 
                  please read our Privacy Policy.
                </p>
                <div className="flex flex-wrap gap-3">
                  <Link to="/privacy">
                    <Button variant="outline" size="sm">
                      Privacy Policy
                    </Button>
                  </Link>
                  <Link to="/settings">
                    <Button variant="outline" size="sm">
                      Back to Settings
                    </Button>
                  </Link>
                </div>
              </Stack>
            </CardBody>
          </Card>
        </div>
      </Container>
    </>
  );
}
