import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';

export function RisingFeedPage() {
  return (
    <>
      <SEO
        title="Rising Clips"
        description="Discover Twitch clips gaining momentum across live culture and catch breakout creators and moments early."
        canonicalUrl="/rising"
      />
      <Container className="py-8">
        <ClipFeed
          title="Rising Clips"
          description="Clips trending upward"
          defaultSort="rising"
          showSearch
        />
      </Container>
    </>
  );
}
