import { Container, Card, CardBody } from '../components';

export function CommunityRulesPage() {
  const lastUpdated = 'January 15, 2025';

  return (
    <Container className="py-8 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">Community Rules</h1>
        <p className="text-sm text-muted-foreground">Last updated: {lastUpdated}</p>
      </div>

      <div className="space-y-6">
        {/* Introduction */}
        <Card>
          <CardBody>
            <p className="text-muted-foreground">
              Welcome to clpr! Our community is built on respect, authenticity, and a shared passion for gaming. 
              These rules help ensure clpr remains a positive space for everyone to discover, share, and celebrate 
              amazing gaming moments.
            </p>
          </CardBody>
        </Card>

        {/* Be Respectful */}
        <Card id="be-respectful">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">1. Be Respectful and Kind</h2>
            <p className="text-muted-foreground mb-4">
              Treat all community members with respect and courtesy. We're here to celebrate gaming, not to tear each other down.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>No harassment, hate speech, or discriminatory language</li>
              <li>Respect differing opinions and viewpoints</li>
              <li>Keep criticism constructive and focused on content, not individuals</li>
              <li>Be welcoming to new community members</li>
            </ul>
          </CardBody>
        </Card>

        {/* Authentic Content */}
        <Card id="authentic-content">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">2. Share Authentic Content</h2>
            <p className="text-muted-foreground mb-4">
              clpr is for genuine gaming highlights. Help maintain quality and authenticity in our community.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Only submit clips from legitimate Twitch streams</li>
              <li>Don't manipulate votes or engagement through bots or fake accounts</li>
              <li>Provide accurate titles and tags for your submissions</li>
              <li>Don't repost content that's already been shared recently</li>
              <li>Give credit where credit is due</li>
            </ul>
          </CardBody>
        </Card>

        {/* No Spam */}
        <Card id="no-spam">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">3. No Spam or Self-Promotion Abuse</h2>
            <p className="text-muted-foreground mb-4">
              Sharing your own content is welcome, but excessive self-promotion disrupts the community experience.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Follow the 80/20 rule: 80% community engagement, 20% self-promotion</li>
              <li>Don't flood the platform with multiple clips in quick succession</li>
              <li>No advertising, referral links, or unrelated promotional content</li>
              <li>Participate genuinely in the community, not just to promote yourself</li>
            </ul>
          </CardBody>
        </Card>

        {/* Safe Content */}
        <Card id="safe-content">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">4. Keep Content Safe and Appropriate</h2>
            <p className="text-muted-foreground mb-4">
              clpr should be accessible to a wide gaming audience. Some content doesn't belong here.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>No NSFW (Not Safe For Work) content</li>
              <li>No graphic violence or disturbing content</li>
              <li>No illegal activities or content that violates Twitch's Terms of Service</li>
              <li>Tag content appropriately for spoilers or sensitive topics</li>
            </ul>
          </CardBody>
        </Card>

        {/* Respect Privacy */}
        <Card id="respect-privacy">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">5. Respect Privacy and Personal Information</h2>
            <p className="text-muted-foreground mb-4">
              Protect your privacy and the privacy of others in the community.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Don't share personal information without consent</li>
              <li>No doxxing or revealing private details about others</li>
              <li>Respect streamers' and creators' boundaries</li>
              <li>Report content that violates someone's privacy</li>
            </ul>
          </CardBody>
        </Card>

        {/* Follow Platform Guidelines */}
        <Card id="platform-guidelines">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">6. Follow Platform and Legal Guidelines</h2>
            <p className="text-muted-foreground mb-4">
              Respect intellectual property and comply with platform rules.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Respect copyright and intellectual property rights</li>
              <li>Don't share pirated content or encourage illegal activities</li>
              <li>Comply with Twitch's Terms of Service and Community Guidelines</li>
              <li>Follow any additional game-specific or publisher guidelines</li>
            </ul>
          </CardBody>
        </Card>

        {/* Report Issues */}
        <Card id="report-issues">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">7. Report Issues and Help Moderate</h2>
            <p className="text-muted-foreground mb-4">
              Help us maintain a healthy community by reporting problems when you see them.
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Use the report feature for content that violates these rules</li>
              <li>Don't engage with trolls or escalate conflicts</li>
              <li>Trust our moderation team to handle reports fairly</li>
              <li>Contact moderators directly for urgent issues</li>
            </ul>
          </CardBody>
        </Card>

        {/* Consequences */}
        <Card id="consequences">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Enforcement and Consequences</h2>
            <p className="text-muted-foreground mb-4">
              Violations of these community rules may result in:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li><strong>Warning:</strong> First-time or minor violations receive a formal warning</li>
              <li><strong>Content Removal:</strong> Violating content will be removed from the platform</li>
              <li><strong>Temporary Suspension:</strong> Repeated violations may result in temporary account suspension</li>
              <li><strong>Permanent Ban:</strong> Serious or repeated violations may result in permanent account termination</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              All moderation decisions are made at the discretion of the clpr team. We review each case individually 
              and aim to be fair and consistent in our enforcement.
            </p>
          </CardBody>
        </Card>

        {/* Contact */}
        <Card id="contact">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Questions or Concerns?</h2>
            <p className="text-muted-foreground mb-4">
              If you have questions about these rules or want to report a concern, you can:
            </p>
            <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
              <li>Open an issue on our <a href="https://git.subcult.tv/subculture-collective/clpr" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">GitHub repository</a></li>
              <li>Contact the moderation team through the platform's report feature</li>
              <li>Join our community Discord for general questions</li>
            </ul>
            <p className="text-muted-foreground mt-4">
              Thank you for being part of the clpr community and helping make it a great place for all gamers! 🎮
            </p>
          </CardBody>
        </Card>
      </div>
    </Container>
  );
}
