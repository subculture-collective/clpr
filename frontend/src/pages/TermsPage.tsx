import { Container, Card, CardBody, SEO } from '../components';
import { Link } from 'react-router-dom';

export function TermsPage() {
  const lastUpdated = 'January 15, 2025';

  return (
    <>
      <SEO
        title="Terms of Service"
        description="Read Clipper's Terms of Service. Learn about eligibility, user conduct, content policies, and your rights and responsibilities when using our platform."
        canonicalUrl="/terms"
      />
      <Container className="py-8 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">Terms of Service</h1>
        <p className="text-sm text-muted-foreground">Last updated: {lastUpdated}</p>
      </div>

      <div className="space-y-6">
        {/* Introduction */}
        <Card>
          <CardBody>
            <p className="text-muted-foreground mb-4">
              Welcome to Clipper! These Terms of Service ("Terms") govern your access to and use of the Clipper 
              platform, including our website, services, and applications (collectively, the "Service").
            </p>
            <p className="text-muted-foreground mb-4">
              By accessing or using Clipper, you agree to be bound by these Terms and our{' '}
              <Link to="/privacy" className="text-primary hover:underline">Privacy Policy</Link>. 
              If you don't agree to these Terms, please don't use our Service.
            </p>
            <p className="text-muted-foreground">
              <strong className="text-foreground">Important:</strong> These Terms contain a binding arbitration 
              clause and class action waiver that affect your rights. Please read them carefully.
            </p>
          </CardBody>
        </Card>

        {/* Eligibility */}
        <Card id="eligibility">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">1. Eligibility</h2>
            <p className="text-muted-foreground mb-4">
              To use Clipper, you must:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Be at least 13 years old</li>
              <li>Have the legal capacity to enter into these Terms</li>
              <li>Not be prohibited from using the Service under applicable laws</li>
              <li>Comply with all local laws and regulations in your jurisdiction</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              If you're under 18, you represent that you have your parent's or guardian's permission to use 
              the Service. We may ask for proof of age or parental consent at any time.
            </p>
          </CardBody>
        </Card>

        {/* Account Registration */}
        <Card id="account-registration">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">2. Account Registration and Security</h2>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Creating an Account</h3>
                <p className="text-muted-foreground mb-2">
                  To access certain features, you must create an account via Twitch OAuth. You agree to:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Provide accurate and complete information</li>
                  <li>Maintain and update your information as needed</li>
                  <li>Keep your account credentials secure</li>
                  <li>Not share your account with others</li>
                  <li>Notify us immediately of any unauthorized access</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Account Responsibility</h3>
                <p className="text-muted-foreground">
                  You are responsible for all activity that occurs under your account. You agree to accept 
                  responsibility for all activities that occur under your account or password.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Acceptable Use */}
        <Card id="acceptable-use">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">3. Acceptable Use Policy</h2>
            <p className="text-muted-foreground mb-4">
              When using Clipper, you agree NOT to:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Violate any laws or regulations</li>
              <li>Infringe on intellectual property rights</li>
              <li>Post harmful, offensive, or inappropriate content</li>
              <li>Harass, threaten, or harm other users</li>
              <li>Impersonate others or misrepresent your affiliation</li>
              <li>Use bots, scrapers, or automated tools without permission</li>
              <li>Attempt to gain unauthorized access to our systems</li>
              <li>Interfere with or disrupt the Service</li>
              <li>Upload viruses, malware, or malicious code</li>
              <li>Manipulate votes, views, or engagement metrics</li>
              <li>Engage in spam or excessive self-promotion</li>
              <li>Collect user data without consent</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              For more details, please review our{' '}
              <Link to="/community-rules" className="text-primary hover:underline">
                Community Rules
              </Link>.
            </p>
          </CardBody>
        </Card>

        {/* Content and Licensing */}
        <Card id="content-licensing">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">4. Content and Intellectual Property</h2>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Your Content</h3>
                <p className="text-muted-foreground mb-2">
                  When you submit content to Clipper (clips, comments, votes, etc.), you grant us a worldwide, 
                  non-exclusive, royalty-free license to use, reproduce, modify, adapt, publish, and display 
                  that content in connection with the Service.
                </p>
                <p className="text-muted-foreground">
                  You retain ownership of your content, but you give us permission to make it available to 
                  other users and to use it to improve and promote the Service.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Your Responsibilities</h3>
                <p className="text-muted-foreground mb-2">
                  You represent and warrant that:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>You own or have the necessary rights to submit the content</li>
                  <li>Your content doesn't violate any third-party rights</li>
                  <li>Your content complies with these Terms and applicable laws</li>
                  <li>You have obtained all necessary permissions for content featuring others</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Our Content</h3>
                <p className="text-muted-foreground">
                  The Service and its original content (excluding user submissions), features, and functionality 
                  are owned by Clipper and are protected by copyright, trademark, and other intellectual property 
                  laws. Our source code is available under an open-source license on GitHub.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Twitch Content</h3>
                <p className="text-muted-foreground">
                  Clips displayed on Clipper are sourced from Twitch and remain subject to Twitch's Terms of 
                  Service. We display clips via embedded players or direct links, respecting creators' rights 
                  and Twitch's platform policies.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* DMCA and Copyright */}
        <Card id="dmca">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">5. Copyright and DMCA</h2>
            <p className="text-muted-foreground mb-4">
              We respect intellectual property rights. If you believe content on Clipper infringes your copyright, 
              please submit a DMCA takedown notice to us via our GitHub repository with:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Identification of the copyrighted work claimed to be infringed</li>
              <li>Identification of the infringing material and its location</li>
              <li>Your contact information</li>
              <li>A statement of good faith belief</li>
              <li>A statement made under penalty of perjury</li>
              <li>Your physical or electronic signature</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              We will respond to valid DMCA notices in accordance with the Digital Millennium Copyright Act.
            </p>
          </CardBody>
        </Card>

        {/* Moderation */}
        <Card id="moderation">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">6. Content Moderation</h2>
            <p className="text-muted-foreground mb-4">
              We reserve the right, but not the obligation, to:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Monitor and review content submitted to the Service</li>
              <li>Remove or modify content that violates these Terms</li>
              <li>Suspend or terminate accounts that violate our policies</li>
              <li>Cooperate with law enforcement when required</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              We are not responsible for user-generated content and don't endorse any opinions expressed by users. 
              However, we take violations of our policies seriously and will take appropriate action.
            </p>
          </CardBody>
        </Card>

        {/* Termination */}
        <Card id="termination">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">7. Termination</h2>
            <p className="text-muted-foreground mb-4">
              We may suspend or terminate your access to the Service at any time, with or without notice, for 
              any reason, including if:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>You violate these Terms or our policies</li>
              <li>We're required to do so by law</li>
              <li>We discontinue the Service or any part of it</li>
              <li>We determine your use poses a security risk</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              You may also delete your account at any time through your account settings. Upon termination, 
              your right to use the Service will immediately cease.
            </p>
          </CardBody>
        </Card>

        {/* Disclaimers */}
        <Card id="disclaimers">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">8. Disclaimers and Limitation of Liability</h2>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Service "As Is"</h3>
                <p className="text-muted-foreground">
                  THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EITHER 
                  EXPRESS OR IMPLIED. WE DISCLAIM ALL WARRANTIES, INCLUDING MERCHANTABILITY, FITNESS FOR A 
                  PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Limitation of Liability</h3>
                <p className="text-muted-foreground mb-2">
                  TO THE MAXIMUM EXTENT PERMITTED BY LAW, CLIPPER SHALL NOT BE LIABLE FOR ANY:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Indirect, incidental, special, or consequential damages</li>
                  <li>Loss of profits, data, or goodwill</li>
                  <li>Service interruptions or errors</li>
                  <li>User content or conduct</li>
                  <li>Third-party services or content</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Indemnification</h3>
                <p className="text-muted-foreground">
                  You agree to indemnify and hold harmless Clipper and its affiliates from any claims, damages, 
                  or expenses arising from your use of the Service, your content, or your violation of these Terms.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Third-Party Services */}
        <Card id="third-party">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">9. Third-Party Services and Links</h2>
            <p className="text-muted-foreground">
              The Service integrates with third-party services (e.g., Twitch) and may contain links to external 
              websites. We are not responsible for the content, privacy practices, or terms of service of these 
              third parties. Your use of third-party services is at your own risk.
            </p>
          </CardBody>
        </Card>

        {/* Subscription Terms */}
        <Card id="subscriptions">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">10. Subscriptions and Payments</h2>
            <p className="text-muted-foreground mb-4">
              If we offer paid subscriptions or features:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Subscription fees are billed in advance on a recurring basis</li>
              <li>Prices are subject to change with 30 days notice</li>
              <li>You may cancel your subscription at any time</li>
              <li>Refunds are provided in accordance with our refund policy</li>
              <li>All payments are processed securely through third-party payment providers</li>
            </ul>
          </CardBody>
        </Card>

        {/* Dispute Resolution */}
        <Card id="dispute-resolution">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">11. Dispute Resolution and Arbitration</h2>
            
            <div className="space-y-3">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Informal Resolution</h3>
                <p className="text-muted-foreground">
                  Before filing a claim, you agree to try to resolve the dispute informally by contacting us 
                  through our GitHub repository. We'll try to resolve it within 60 days.
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Binding Arbitration</h3>
                <p className="text-muted-foreground mb-2">
                  If we can't resolve the dispute informally, you agree that all disputes will be resolved 
                  through binding arbitration rather than in court, except for:
                </p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                  <li>Small claims court actions</li>
                  <li>Intellectual property disputes</li>
                  <li>Claims for injunctive relief</li>
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Class Action Waiver</h3>
                <p className="text-muted-foreground">
                  You agree that disputes will be resolved on an individual basis only, not as a class action, 
                  consolidated action, or representative action.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Changes to Terms */}
        <Card id="changes">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">12. Changes to These Terms</h2>
            <p className="text-muted-foreground">
              We may update these Terms from time to time. We'll notify you of material changes by updating 
              the "Last updated" date and, for significant changes, through the Service or via email. Your 
              continued use of the Service after changes become effective constitutes acceptance of the new Terms.
            </p>
          </CardBody>
        </Card>

        {/* General Provisions */}
        <Card id="general">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">13. General Provisions</h2>
            <ul className="space-y-3 text-muted-foreground">
              <li>
                <strong className="text-foreground">Entire Agreement:</strong> These Terms constitute the 
                entire agreement between you and Clipper regarding the Service.
              </li>
              <li>
                <strong className="text-foreground">Severability:</strong> If any provision is found invalid, 
                the remaining provisions will remain in effect.
              </li>
              <li>
                <strong className="text-foreground">Waiver:</strong> Our failure to enforce any right or 
                provision doesn't constitute a waiver.
              </li>
              <li>
                <strong className="text-foreground">Assignment:</strong> You may not assign these Terms without 
                our consent. We may assign them without restriction.
              </li>
              <li>
                <strong className="text-foreground">Governing Law:</strong> These Terms are governed by the 
                laws of the jurisdiction where Clipper is based.
              </li>
            </ul>
          </CardBody>
        </Card>

        {/* Contact */}
        <Card id="contact">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">14. Contact Information</h2>
            <p className="text-muted-foreground mb-4">
              If you have questions about these Terms, please contact us:
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
                <strong className="text-foreground">Email:</strong> legal@clpr.com (for legal inquiries)
              </li>
            </ul>
            <p className="text-muted-foreground mt-4">
              Thank you for using Clipper! We're excited to have you as part of our gaming community. 🎮
            </p>
          </CardBody>
        </Card>
      </div>
    </Container>
    </>
  );
}
