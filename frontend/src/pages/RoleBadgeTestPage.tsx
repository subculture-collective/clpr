import { Container, Card, UserRoleBadge } from '../components';

/**
 * Test page to demonstrate UserRoleBadge component
 * This is for development/testing purposes only
 */
export function RoleBadgeTestPage() {
  return (
    <Container className="py-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">User Role Badge Examples</h1>
        
        <Card className="p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Role Badges (Small)</h2>
          <div className="flex flex-wrap gap-3">
            <UserRoleBadge role="admin" size="sm" />
            <UserRoleBadge role="moderator" size="sm" />
            <UserRoleBadge role="user" size="sm" />
            <UserRoleBadge role="creator" size="sm" />
            <UserRoleBadge role="member" size="sm" />
          </div>
        </Card>

        <Card className="p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Role Badges (Medium)</h2>
          <div className="flex flex-wrap gap-3">
            <UserRoleBadge role="admin" size="md" />
            <UserRoleBadge role="moderator" size="md" />
            <UserRoleBadge role="user" size="md" />
            <UserRoleBadge role="creator" size="md" />
            <UserRoleBadge role="member" size="md" />
          </div>
        </Card>

        <Card className="p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Role Badges (Large)</h2>
          <div className="flex flex-wrap gap-3">
            <UserRoleBadge role="admin" size="lg" />
            <UserRoleBadge role="moderator" size="lg" />
            <UserRoleBadge role="user" size="lg" />
            <UserRoleBadge role="creator" size="lg" />
            <UserRoleBadge role="member" size="lg" />
          </div>
        </Card>

        <Card className="p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">In Context Examples</h2>
          
          <div className="space-y-4">
            <div className="border-b border-border pb-4">
              <h3 className="font-semibold mb-2">Profile Header</h3>
              <div className="flex items-center gap-2">
                <img
                  src="https://via.placeholder.com/40"
                  alt="User avatar"
                  className="w-10 h-10 rounded-full"
                />
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">AdminUser123</span>
                    <UserRoleBadge role="admin" size="sm" />
                  </div>
                  <p className="text-sm text-muted-foreground">1,234 uppies</p>
                </div>
              </div>
            </div>

            <div className="border-b border-border pb-4">
              <h3 className="font-semibold mb-2">Comment Header</h3>
              <div className="flex items-center gap-2">
                <img
                  src="https://via.placeholder.com/32"
                  alt="User avatar"
                  className="w-8 h-8 rounded-full"
                />
                <span className="font-medium">ModUser456</span>
                <UserRoleBadge role="moderator" size="sm" />
                <span className="text-xs text-muted-foreground">500 uppies</span>
                <span className="text-xs text-muted-foreground">•</span>
                <span className="text-xs text-muted-foreground">2 hours ago</span>
              </div>
            </div>

            <div className="border-b border-border pb-4">
              <h3 className="font-semibold mb-2">Submission Author</h3>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Submitted by</span>
                <span className="font-medium">RegularUser789</span>
                <UserRoleBadge role="user" size="sm" />
                <span className="text-xs text-muted-foreground">100 uppies</span>
              </div>
            </div>

            <div>
              <h3 className="font-semibold mb-2">Content Creator</h3>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Created by</span>
                <span className="font-medium">ContentCreator</span>
                <UserRoleBadge role="creator" size="sm" />
              </div>
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Tooltip Information</h2>
          <p className="text-sm text-muted-foreground mb-4">
            Hover over each badge to see the tooltip with role permissions:
          </p>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <UserRoleBadge role="admin" size="sm" />
              <span className="text-sm">- Full access to all platform features and moderation tools</span>
            </div>
            <div className="flex items-center gap-2">
              <UserRoleBadge role="moderator" size="sm" />
              <span className="text-sm">- Can review submissions, manage content, and enforce community guidelines</span>
            </div>
            <div className="flex items-center gap-2">
              <UserRoleBadge role="creator" size="sm" />
              <span className="text-sm">- Created or submitted this clip</span>
            </div>
            <div className="flex items-center gap-2">
              <UserRoleBadge role="user" size="sm" />
              <span className="text-sm">- Regular community member</span>
            </div>
          </div>
        </Card>
      </div>
    </Container>
  );
}
