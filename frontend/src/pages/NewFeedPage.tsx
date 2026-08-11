import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';

export function NewFeedPage() {
  return (
    <>
      <SEO
        title="New Clips"
        description="Discover the latest Twitch clips from the community and fresh moments from creators across every kind of stream."
        canonicalUrl="/new"
      />
      <Container className="py-8">
        <ClipFeed
          title="New Clips"
          description="Latest clips from the community"
          defaultSort="new"
          showSearch
        />
      </Container>
    </>
  );
}
