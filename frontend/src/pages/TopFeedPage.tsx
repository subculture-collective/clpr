import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';

export function TopFeedPage() {
  return (
    <>
      <SEO
        title="Top Clips"
        description="Watch the highest-rated Twitch clips across creators, topics, and communities, as voted by the people watching."
        canonicalUrl="/top"
      />
      <Container className="py-8">
        <ClipFeed
          title="Top Clips"
          description="Top rated clips"
          defaultSort="top"
          defaultTimeframe="day"
          showSearch
        />
      </Container>
    </>
  );
}
