import { useParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';

export function TagPage() {
  const { tagSlug } = useParams<{ tagSlug: string }>();

  if (!tagSlug) {
    return (
      <Container className="py-8">
        <div className="text-center text-muted-foreground">
          <p>No tag specified</p>
        </div>
      </Container>
    );
  }

  return (
    <>
      <SEO
        title={`#${tagSlug}`}
        description={`Explore Twitch clips tagged with ${tagSlug}. Watch, vote, and discover the best ${tagSlug} moments from the community.`}
        canonicalUrl={`/tags/${tagSlug}`}
      />
      <Container className="py-8">
        <ClipFeed
          title={`#${tagSlug}`}
          description={`Clips tagged with ${tagSlug}`}
          filters={{ tags: [tagSlug] }}
          showSearch={false}
          useSortTitle={false}
        />
      </Container>
    </>
  );
}
