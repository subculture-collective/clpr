import { Container, Card, CardBody, SEO } from '../components';
import { Link } from 'react-router-dom';

export function AboutPage() {
  const lastUpdated = 'January 15, 2025';

  return (
    <>
      <SEO
        title="About"
        description="Learn about clpr - a modern, open-source platform for discovering and sharing the best Twitch clips. Join our community of gamers and streamers."
        canonicalUrl="/about"
      />
      <Container className="py-8 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">About clpr</h1>
        <p className="text-sm text-muted-foreground">Last updated: {lastUpdated}</p>
      </div>

      <div className="space-y-6">
        {/* What is clpr */}
        <Card id="what-is-clpr">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">What is clpr?</h2>
            <p className="text-muted-foreground mb-4">
              clpr is a modern, open-source platform for discovering and sharing gaming highlights from Twitch. 
              We aggregate the best clips from your favorite games and streamers, making them easy to find, watch, 
              and share with the gaming community.
            </p>
            <p className="text-muted-foreground mb-4">
              Our mission is to celebrate great gaming moments and connect the gaming community through shared experiences. 
              Whether you're looking for the latest esports plays, hilarious streamer reactions, or incredible speedrun achievements, 
              clpr brings it all together in one place.
            </p>
            <p className="text-muted-foreground">
              Built with React, TypeScript, and modern web technologies, clpr is designed to be fast, responsive, 
              and accessible on any device.
            </p>
          </CardBody>
        </Card>

        {/* How It Works */}
        <Card id="how-it-works">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">How It Works</h2>
            <div className="space-y-4">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Automated Clip Discovery</h3>
                <p className="text-muted-foreground">
                  We automatically sync and index clips from Twitch, ensuring you never miss the hottest moments 
                  from your favorite games and creators.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Smart Browsing & Search</h3>
                <p className="text-muted-foreground">
                  Browse clips by game, creator, or tag. Our intelligent search makes it easy to find exactly 
                  what you're looking for, from specific plays to trending moments.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Community-Driven Curation</h3>
                <p className="text-muted-foreground">
                  Save your favorite clips, upvote the best moments, and see what's trending in the community. 
                  Our feeds (New, Top, Rising) help surface the content that matters most.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Creator Features</h3>
                <p className="text-muted-foreground">
                  Track your clips' performance with analytics, build your audience, and connect with fans 
                  who appreciate your best moments.
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Features */}
        <Card id="features">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Key Features</h2>
            <ul className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Browse clips from multiple games and creators</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Advanced search and filtering options</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Save favorites for later viewing</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Upvote and comment on clips</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Submit your own Twitch clips</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Track trending and rising content</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Creator analytics and insights</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Dark mode and responsive design</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Leaderboards and community stats</span>
              </li>
              <li className="flex items-start">
                <span className="text-primary mr-2">✓</span>
                <span className="text-muted-foreground">Notification system for updates</span>
              </li>
            </ul>
          </CardBody>
        </Card>

        {/* Open Source */}
        <Card id="open-source">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Open Source & Community</h2>
            <p className="text-muted-foreground mb-4">
              clpr is proudly open source! Our code is available on GitHub, and we welcome contributions 
              from developers, designers, and gaming enthusiasts.
            </p>
            <div className="flex flex-wrap gap-4">
              <a
                href="https://git.subcult.tv/subculture-collective/clpr"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
              >
                View on GitHub
              </a>
              <a
                href="https://git.subcult.tv/subculture-collective/clpr/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center px-4 py-2 border border-border rounded-md hover:bg-accent transition-colors"
              >
                Report Issues
              </a>
              <a
                href="https://git.subcult.tv/subculture-collective/clpr/blob/main/CONTRIBUTING.md"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center px-4 py-2 border border-border rounded-md hover:bg-accent transition-colors"
              >
                Contribute
              </a>
            </div>
          </CardBody>
        </Card>

        {/* Technology Stack */}
        <Card id="tech-stack">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Technology Stack</h2>
            <p className="text-muted-foreground mb-4">
              clpr is built with modern, production-ready technologies:
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Frontend</h3>
                <ul className="space-y-1 text-muted-foreground">
                  <li>• React 19 with TypeScript</li>
                  <li>• Vite for build tooling</li>
                  <li>• TailwindCSS for styling</li>
                  <li>• React Router for navigation</li>
                  <li>• TanStack Query for data fetching</li>
                </ul>
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-2 text-foreground">Backend</h3>
                <ul className="space-y-1 text-muted-foreground">
                  <li>• Python with FastAPI</li>
                  <li>• PostgreSQL database</li>
                  <li>• Redis for caching</li>
                  <li>• Twitch API integration</li>
                  <li>• Docker for deployment</li>
                </ul>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Contact */}
        <Card id="contact">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-4">Get in Touch</h2>
            <p className="text-muted-foreground mb-4">
              Have questions, feedback, or just want to connect? We'd love to hear from you!
            </p>
            <div className="space-y-2 text-muted-foreground">
              <p>
                <strong className="text-foreground">Community:</strong> Join our{' '}
                <Link to="/community-rules" className="text-primary hover:underline">
                  community
                </Link>{' '}
                and follow our guidelines
              </p>
              <p>
                <strong className="text-foreground">Development:</strong> Contribute on{' '}
                <a
                  href="https://git.subcult.tv/subculture-collective/clpr"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  GitHub
                </a>
              </p>
              <p>
                <strong className="text-foreground">Legal:</strong> Review our{' '}
                <Link to="/privacy" className="text-primary hover:underline">
                  Privacy Policy
                </Link>{' '}
                and{' '}
                <Link to="/terms" className="text-primary hover:underline">
                  Terms of Service
                </Link>
              </p>
            </div>
          </CardBody>
        </Card>
      </div>
    </Container>
    </>
  );
}
