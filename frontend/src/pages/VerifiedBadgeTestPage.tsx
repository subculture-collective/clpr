
import { Container, Card, CardBody } from '@/components';
import { VerifiedBadge } from '@/components/user';

/**
 * Test page to showcase the VerifiedBadge component
 * This is used for visual testing and screenshots
 */
export function VerifiedBadgeTestPage() {
  return (
    <Container className="py-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <h1 className="text-3xl font-bold mb-6">Verification Badge Test Page</h1>

        {/* Badge Sizes */}
        <Card>
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">Badge Sizes</h2>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Small:</span>
                <VerifiedBadge size="sm" />
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Medium:</span>
                <VerifiedBadge size="md" />
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Large:</span>
                <VerifiedBadge size="lg" />
              </div>
            </div>
          </CardBody>
        </Card>

        {/* In Context - User Profile Header */}
        <Card>
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">In Context: Profile Header</h2>
            <div className="flex items-start gap-4">
              <img
                src="https://via.placeholder.com/96"
                alt="Example user avatar for demonstration"
                className="w-24 h-24 rounded-full border-2 border-border"
              />
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="text-2xl font-bold">John Verified User</h3>
                  <VerifiedBadge size="lg" />
                </div>
                <p className="text-muted-foreground mb-2">@johnverified</p>
                <p className="text-foreground">This user is verified by Clipper administrators.</p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* In Context - Comment Thread */}
        <Card>
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">In Context: Comment Thread</h2>
            <div className="space-y-4">
              <div className="flex gap-3">
                <img
                  src="https://via.placeholder.com/40"
                  alt="Example comment author avatar for demonstration"
                  className="w-10 h-10 rounded-full"
                />
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium">VerifiedCreator</span>
                    <VerifiedBadge size="sm" />
                    <span className="text-xs text-muted-foreground">• 2 hours ago</span>
                  </div>
                  <p className="text-foreground">
                    This is a comment from a verified creator. The badge helps users identify authentic accounts.
                  </p>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* In Context - Clip Submission */}
        <Card>
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">In Context: Clip Submission</h2>
            <div className="text-sm text-muted-foreground">
              <span>Clipped by </span>
              <a href="#" className="hover:text-foreground">StreamerName</a>
              <span> • Submitted by </span>
              <a href="#" className="hover:text-foreground inline-flex items-center gap-1">
                VerifiedUser
                <VerifiedBadge size="sm" />
              </a>
              <span> • 3 hours ago</span>
            </div>
          </CardBody>
        </Card>

        {/* Tooltip Demo */}
        <Card>
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">Tooltip</h2>
            <p className="text-muted-foreground mb-4">
              Hover over the badge to see the tooltip explanation:
            </p>
            <div className="flex items-center gap-2">
              <span className="font-medium">Hover me:</span>
              <VerifiedBadge size="lg" />
            </div>
          </CardBody>
        </Card>

        {/* Dark Mode Preview */}
        <Card className="bg-background text-white">
          <CardBody>
            <h2 className="text-xl font-semibold mb-4">Dark Mode</h2>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <span className="font-medium">Verified User</span>
                <VerifiedBadge size="md" />
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">With smaller badge:</span>
                <VerifiedBadge size="sm" />
              </div>
            </div>
          </CardBody>
        </Card>
      </div>
    </Container>
  );
}
