import { Container, Card, CardBody, SEO } from '../components';

export function PrivacyPage() {
  const lastUpdated = 'January 15, 2025';

  return (
    <>
      <SEO
        title="Privacy Policy"
        description="Learn how clpr collects, uses, and protects your personal information. Read our Privacy Policy to understand your data rights and our privacy practices."
        canonicalUrl="/privacy"
      />
      <Container className="py-8 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">Privacy Policy</h1>
        <p className="text-sm text-muted-foreground">Last updated: {lastUpdated}</p>
      </div>

      <div className="space-y-6">
        {/* Introduction */}
        <Card>
          <CardBody>
            <p className="text-muted-foreground mb-4">
              At clpr, we take your privacy seriously. This Privacy Policy explains how we collect, use, 
              disclose, and safeguard your information when you use our platform. Please read this policy 
              carefully to understand our practices regarding your personal data.
            </p>
            <p className="text-muted-foreground">
              By using clpr, you agree to the collection and use of information in accordance with this policy. 
              If you do not agree with our policies and practices, please do not use our platform.
            </p>
          </CardBody>
        </Card>

        {/* Information We Collect */}
        <Card id="information-we-collect">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Information We Collect</h2>
            
            <div className="space-y-4">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Information You Provide</h3>
                <p className="text-muted-foreground mb-2">
                  When you create an account or use our services, we collect:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Twitch account information (username, profile picture) via OAuth</li>
                  <li>Email address (if provided through Twitch)</li>
                  <li>Content you submit (clips, comments, votes)</li>
                  <li>Settings and preferences</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Automatically Collected Information</h3>
                <p className="text-muted-foreground mb-2">
                  When you access clpr, we automatically collect:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Device information (browser type, operating system)</li>
                  <li>IP address and general location data</li>
                  <li>Usage data (pages visited, time spent, interactions)</li>
                  <li>Cookies and similar tracking technologies</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Third-Party Information</h3>
                <p className="text-muted-foreground mb-2">
                  We receive information from third-party services:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Twitch API (publicly available clip and stream data)</li>
                  <li>OAuth providers for authentication</li>
                  <li>Analytics services (when enabled)</li>
                </ul>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* How We Use Information */}
        <Card id="how-we-use-information">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">How We Use Your Information</h2>
            <p className="text-muted-foreground mb-4">
              We use the information we collect to:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Provide, maintain, and improve our services</li>
              <li>Authenticate your account and manage your profile</li>
              <li>Display your submissions, comments, and votes</li>
              <li>Personalize your experience and show relevant content</li>
              <li>Send notifications about your account and community activity</li>
              <li>Analyze usage patterns to improve platform performance</li>
              <li>Detect, prevent, and address technical issues and abuse</li>
              <li>Comply with legal obligations and enforce our terms</li>
              <li>Communicate with you about updates and features</li>
            </ul>
          </CardBody>
        </Card>

        {/* Cookies and Tracking */}
        <Card id="cookies-tracking">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Cookies and Tracking Technologies</h2>
            <p className="text-muted-foreground mb-4">
              We use cookies and similar tracking technologies to track activity on our platform and store 
              certain information. Cookies are small data files stored on your device.
            </p>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Essential Cookies</h3>
                <p className="text-muted-foreground">
                  Required for authentication, security, and core platform functionality. These cannot be disabled.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Preference Cookies</h3>
                <p className="text-muted-foreground">
                  Remember your settings like theme preference and language selection.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Analytics Cookies</h3>
                <p className="text-muted-foreground">
                  Help us understand how you use clpr so we can improve the platform. These are optional 
                  and can be disabled in your browser settings.
                </p>
              </div>
            </div>

            <p className="text-muted-foreground mt-4">
              You can instruct your browser to refuse all cookies or indicate when a cookie is being sent. 
              However, some features may not function properly without cookies.
            </p>
          </CardBody>
        </Card>

        {/* Advertising and Personalization */}
        <Card id="advertising">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Advertising and Personalization</h2>
            <p className="text-muted-foreground mb-4">
              clpr displays advertisements to support our free service. We offer you control over how 
              ads are personalized based on your preferences and behavior.
            </p>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Contextual Advertising</h3>
                <p className="text-muted-foreground">
                  By default, you will see contextual ads based on the content you're viewing (such as the 
                  game category or page content). This type of advertising does not require tracking your 
                  personal browsing behavior.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Personalized Advertising</h3>
                <p className="text-muted-foreground">
                  With your consent, we may show personalized ads based on your interests, viewing history, 
                  and demographics. This includes using information such as your country, device type, and 
                  content preferences to display more relevant advertisements.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Your Advertising Choices</h3>
                <p className="text-muted-foreground mb-2">
                  You have full control over advertising personalization:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Use the consent banner when you first visit to set your preferences</li>
                  <li>Visit your Settings page at any time to update your consent choices</li>
                  <li>Enable "Do Not Track" in your browser to automatically opt out of personalization</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Do Not Track</h3>
                <p className="text-muted-foreground">
                  We honor the Do Not Track (DNT) and Global Privacy Control (GPC) browser signals. When 
                  enabled, we will not use your personal information for ad personalization, even if you 
                  have previously consented. You will continue to see contextual ads based on page content.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Data Sharing */}
        <Card id="data-sharing">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">How We Share Your Information</h2>
            <p className="text-muted-foreground mb-4">
              We do not sell your personal information. We may share your information in the following situations:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li><strong className="text-foreground">Public Content:</strong> Clips, comments, and votes you submit are publicly visible</li>
              <li><strong className="text-foreground">Service Providers:</strong> With third-party vendors who help us operate the platform (hosting, analytics, error tracking)</li>
              <li><strong className="text-foreground">Legal Requirements:</strong> When required by law or to protect our rights and safety</li>
              <li><strong className="text-foreground">Business Transfers:</strong> In connection with any merger, sale, or acquisition of our assets</li>
              <li><strong className="text-foreground">With Your Consent:</strong> When you explicitly agree to share information</li>
            </ul>
          </CardBody>
        </Card>

        {/* Data Security */}
        <Card id="data-security">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Data Security</h2>
            <p className="text-muted-foreground mb-4">
              We implement appropriate technical and organizational measures to protect your personal information:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Encrypted connections (HTTPS/TLS)</li>
              <li>Secure authentication via OAuth</li>
              <li>Regular security audits and updates</li>
              <li>Access controls and monitoring</li>
              <li>Secure data storage and backups</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              However, no method of transmission over the Internet is 100% secure. While we strive to protect 
              your personal information, we cannot guarantee absolute security.
            </p>
          </CardBody>
        </Card>

        {/* Data Retention */}
        <Card id="data-retention">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Data Retention</h2>
            <p className="text-muted-foreground mb-4">
              We retain your personal information for as long as necessary to:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Provide our services and maintain your account</li>
              <li>Comply with legal obligations</li>
              <li>Resolve disputes and enforce agreements</li>
              <li>Prevent fraud and abuse</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              When you delete your account, we will delete or anonymize your personal information, except where 
              we're required to retain it by law or for legitimate business purposes (e.g., preventing abuse).
            </p>
          </CardBody>
        </Card>

        {/* Your Rights */}
        <Card id="your-rights">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Your Privacy Rights</h2>
            <p className="text-muted-foreground mb-4">
              Depending on your location, you may have the following rights regarding your personal information:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li><strong className="text-foreground">Access:</strong> Request a copy of the personal data we hold about you</li>
              <li><strong className="text-foreground">Correction:</strong> Update or correct inaccurate information</li>
              <li><strong className="text-foreground">Deletion:</strong> Request deletion of your personal data</li>
              <li><strong className="text-foreground">Portability:</strong> Receive your data in a structured, machine-readable format</li>
              <li><strong className="text-foreground">Objection:</strong> Object to certain processing of your data</li>
              <li><strong className="text-foreground">Restriction:</strong> Request restriction of processing in certain circumstances</li>
              <li><strong className="text-foreground">Withdraw Consent:</strong> Withdraw consent where we rely on it for processing</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              To exercise these rights, please contact us through our GitHub repository or settings page. 
              We will respond to your request within 30 days.
            </p>
          </CardBody>
        </Card>

        {/* Third-Party Services */}
        <Card id="third-party-services">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Third-Party Services</h2>
            <p className="text-muted-foreground mb-4">
              Our platform integrates with third-party services that have their own privacy policies:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li><strong className="text-foreground">Twitch:</strong> For authentication and clip data (see{' '}
                <a href="https://www.twitch.tv/p/legal/privacy-notice/" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                  Twitch Privacy Policy
                </a>)
              </li>
              <li><strong className="text-foreground">Error Tracking:</strong> Sentry (when enabled) for error monitoring</li>
              <li><strong className="text-foreground">Payment Processing:</strong> Stripe for subscription payments (if applicable)</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              We are not responsible for the privacy practices of these third parties. We encourage you to 
              review their privacy policies.
            </p>
          </CardBody>
        </Card>

        {/* Children's Privacy */}
        <Card id="childrens-privacy">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Children's Privacy</h2>
            <p className="text-muted-foreground">
              clpr is not intended for users under the age of 13. We do not knowingly collect personal 
              information from children under 13. If you believe we have collected information from a child 
              under 13, please contact us immediately, and we will take steps to delete such information.
            </p>
          </CardBody>
        </Card>

        {/* International Users */}
        <Card id="international-users">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">International Data Transfers</h2>
            <p className="text-muted-foreground">
              Your information may be transferred to and processed in countries other than your own. These 
              countries may have different data protection laws. By using clpr, you consent to the transfer 
              of your information to our servers and third-party service providers, wherever they may be located.
            </p>
          </CardBody>
        </Card>

        {/* Changes to Policy */}
        <Card id="changes">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Changes to This Privacy Policy</h2>
            <p className="text-muted-foreground">
              We may update this Privacy Policy from time to time. We will notify you of any changes by 
              updating the "Last updated" date at the top of this policy. Significant changes will be 
              communicated through the platform or via email. We encourage you to review this policy 
              periodically to stay informed about how we protect your information.
            </p>
          </CardBody>
        </Card>

        {/* Contact */}
        <Card id="contact">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Contact Us</h2>
            <p className="text-muted-foreground mb-4">
              If you have questions or concerns about this Privacy Policy or our data practices, please contact us:
            </p>
            <ul className="space-y-2 text-muted-foreground ml-4">
              <li>
                <strong className="text-foreground">GitHub:</strong>{' '}
                <a
                  href="https://git.subcult.tv/subculture-collective/clpr/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  Open an issue
                </a>
              </li>
              <li>
                <strong className="text-foreground">Email:</strong> privacy@clpr.com (for privacy-specific inquiries)
              </li>
            </ul>
          </CardBody>
        </Card>
      </div>
    </Container>
    </>
  );
}
