import { Link } from 'react-router-dom';
import { Card, CardBody, Container, SEO } from '../components';

const features = [
  'Discover creators across topics and communities',
  'Browse curated collections of standout moments',
  'Save favorites for later viewing',
  'Give clips uppies and join the conversation',
  'Submit Twitch clips worth sharing',
  'Follow what is trending, rising, and new',
];

export function AboutPage() {
  return (
    <>
      <SEO title='About' description='clpr helps people discover the creators and moments shaping live culture.' canonicalUrl='/about' />
      <Container className='py-8 max-w-4xl'>
        <div className='mb-8'>
          <p className='text-sm font-semibold uppercase tracking-widest text-primary mb-3'>Live culture, clipped</p>
          <h1 className='text-4xl font-bold mb-4'>Find the creators everyone will be talking about.</h1>
          <p className='text-lg text-muted-foreground max-w-3xl'>clpr brings together memorable Twitch moments so you can discover people worth following—not just whatever category happens to be live.</p>
        </div>

        <div className='space-y-6'>
          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>More than gaming</h2>
            <p className='text-muted-foreground mb-4'>Live creators move freely between IRL, reactions, music, news, politics, sports, art, gaming, and conversations that do not fit neatly into a box. clpr is built around those creators and the moments their communities remember.</p>
            <p className='text-muted-foreground'>Browse by creator, topic, tag, or Twitch category. Curated collections bring related clips together, while community activity helps timely moments rise.</p>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>What you can do</h2>
            <ul className='grid grid-cols-1 md:grid-cols-2 gap-3'>
              {features.map(feature => (
                <li key={feature} className='flex items-start'>
                  <span aria-hidden='true' className='text-primary mr-2'>✓</span>
                  <span className='text-muted-foreground'>{feature}</span>
                </li>
              ))}
            </ul>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Made for discovery</h2>
            <div className='grid gap-5 md:grid-cols-3'>
              <div><h3 className='font-semibold mb-1'>Creator-first</h3><p className='text-sm text-muted-foreground'>The person behind the moment stays at the center, with a clear path back to their Twitch presence.</p></div>
              <div><h3 className='font-semibold mb-1'>Community-shaped</h3><p className='text-sm text-muted-foreground'>Uppies, saves, comments, and viewing activity help surface clips people genuinely care about.</p></div>
              <div><h3 className='font-semibold mb-1'>Broad by design</h3><p className='text-sm text-muted-foreground'>Topics and tags connect moments across categories instead of treating every stream like a single subject.</p></div>
            </div>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Be part of it</h2>
            <p className='text-muted-foreground mb-4'>Watch, save, vote, share, and submit the moments that deserve a wider audience. Please read our <Link to='/community-rules' className='text-primary hover:underline'>community rules</Link> and help keep clpr welcoming to creators and viewers alike.</p>
            <div className='flex flex-wrap gap-4'>
              <Link to='/contact' className='inline-flex px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90'>Contact us</Link>
              <a href='https://patreon.com/subcult' target='_blank' rel='noopener noreferrer' className='inline-flex px-4 py-2 border border-border rounded-md hover:bg-accent'>Support us on Patreon</a>
            </div>
          </CardBody></Card>
        </div>
      </Container>
    </>
  );
}
