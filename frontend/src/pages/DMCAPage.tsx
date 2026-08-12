import { Card, CardBody, Container, SEO } from '../components';

export function DMCAPage() {
  return (
    <>
      <SEO title='DMCA Copyright Policy' description='How to send clpr a copyright notice or counter-notice.' canonicalUrl='/legal/dmca' />
      <Container className='py-8 max-w-4xl'>
        <div className='mb-8'><h1 className='text-4xl font-bold mb-4'>DMCA Copyright Policy</h1><p className='text-sm text-muted-foreground'>Last updated: August 11, 2026</p></div>
        <div className='space-y-6'>
          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Copyright reports</h2>
            <p className='text-muted-foreground'>clpr respects copyright and responds to complete notices concerning material available through the service. Before sending a notice, consider whether the use is authorized by the copyright owner or permitted by law. You may wish to consult an attorney.</p>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Send a takedown notice</h2>
            <p className='text-muted-foreground mb-4'>Email your notice to <a href='mailto:dmca@clpr.tv' className='text-primary hover:underline'>dmca@clpr.tv</a>. A complete notice should include:</p>
            <ul className='list-disc list-inside space-y-2 text-muted-foreground ml-4'>
              <li>Your physical or electronic signature.</li>
              <li>Identification of the copyrighted work, or a representative list if several works on one site are covered.</li>
              <li>Identification and location of the material you want removed, including the clpr URL.</li>
              <li>Your name, mailing address, telephone number, and email address.</li>
              <li>A statement that you have a good-faith belief the disputed use is not authorized by the copyright owner, its agent, or the law.</li>
              <li>A statement, under penalty of perjury, that the notice is accurate and that you are the copyright owner or authorized to act for the owner.</li>
            </ul>
            <p className='text-muted-foreground mt-4'>Incomplete notices may delay our response. We may share the notice with the person who submitted the material and with transparency-reporting services when appropriate.</p>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Counter-notices</h2>
            <p className='text-muted-foreground mb-4'>If material you submitted was removed because of mistake or misidentification, you may email a counter-notice to <a href='mailto:dmca@clpr.tv' className='text-primary hover:underline'>dmca@clpr.tv</a> containing:</p>
            <ul className='list-disc list-inside space-y-2 text-muted-foreground ml-4'>
              <li>Your physical or electronic signature.</li>
              <li>Identification of the removed material and where it appeared before removal.</li>
              <li>A statement under penalty of perjury that you have a good-faith belief the material was removed or disabled because of mistake or misidentification.</li>
              <li>Your name, mailing address, telephone number, and email address.</li>
              <li>Your consent to the jurisdiction of the appropriate U.S. federal district court and agreement to accept service of process from the original complainant or their agent.</li>
            </ul>
            <p className='text-muted-foreground mt-4'>We may restore material 10–14 business days after forwarding a valid counter-notice unless the complainant tells us they have filed a court action seeking to restrain the activity.</p>
          </CardBody></Card>

          <Card><CardBody>
            <h2 className='text-2xl font-semibold mb-4'>Repeat infringement and misrepresentation</h2>
            <p className='text-muted-foreground mb-3'>When appropriate, we may terminate accounts of repeat infringers. Knowingly making a material misrepresentation in a notice or counter-notice may create liability under 17 U.S.C. § 512(f).</p>
            <p className='text-muted-foreground'>DMCA notices are legal documents. For non-copyright reports, use the report control on the relevant content or contact <a href='mailto:support@clpr.tv' className='text-primary hover:underline'>support@clpr.tv</a>.</p>
          </CardBody></Card>
        </div>
      </Container>
    </>
  );
}
