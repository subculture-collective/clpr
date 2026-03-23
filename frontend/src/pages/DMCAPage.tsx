import { AlertTriangle } from 'lucide-react';
import { Container, Card, CardBody, SEO } from '../components';

export function DMCAPage() {
  const lastUpdated = 'December 12, 2025';

  return (
    <>
      <SEO
        title="DMCA Copyright Policy"
        description="Learn about clpr's DMCA copyright policy, how to file a takedown notice, counter-notice procedures, and our repeat infringer policy."
        canonicalUrl="/legal/dmca"
      />
      <Container className="py-8 max-w-4xl">
        <div className="mb-8">
          <h1 className="text-4xl font-bold mb-4">DMCA Copyright Policy</h1>
          <p className="text-sm text-muted-foreground">Last updated: {lastUpdated}</p>
        </div>

        <div className="space-y-6">
          {/* Overview */}
          <Card>
            <CardBody>
              <p className="text-muted-foreground mb-4">
                Clipper respects intellectual property rights and complies with the Digital Millennium 
                Copyright Act ("DMCA"), 17 U.S.C. § 512. This policy outlines our procedures for handling 
                copyright infringement claims and counter-notices.
              </p>
              <p className="text-muted-foreground">
                Clipper qualifies for DMCA safe harbor protection as an online service provider hosting 
                user-submitted content. We respond expeditiously to remove or disable access to allegedly 
                infringing material upon receiving proper notice.
              </p>
            </CardBody>
          </Card>

          {/* Designated DMCA Agent */}
          <Card id="dmca-agent">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Designated DMCA Agent</h2>
              <p className="text-muted-foreground mb-4">
                All DMCA notices, counter-notices, and copyright-related inquiries must be sent to our 
                designated DMCA agent:
              </p>
              <div className="bg-muted/50 p-4 rounded-lg space-y-2">
                <p className="text-foreground"><strong>Designated Agent:</strong> [Agent Name]</p>
                <p className="text-foreground"><strong>Service Provider:</strong> Subculture Collective (Clipper)</p>
                <p className="text-foreground">
                  <strong>Physical Address:</strong><br />
                  [Street Address]<br />
                  [City, State ZIP Code]<br />
                  United States
                </p>
                <p className="text-foreground"><strong>Email:</strong> dmca@clpr.tv</p>
                <p className="text-foreground"><strong>Phone:</strong> [Phone Number]</p>
              </div>
              <p className="text-muted-foreground mt-4 text-sm">
                <strong>Note:</strong> This agent is designated solely for receiving DMCA notices and 
                counter-notices. For general inquiries, please contact support@clpr.tv.
              </p>
            </CardBody>
          </Card>

          {/* Filing a DMCA Takedown Notice */}
          <Card id="filing-takedown">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Filing a DMCA Takedown Notice</h2>
              <p className="text-muted-foreground mb-4">
                If you believe your copyrighted work has been infringed on Clipper, you may submit a 
                DMCA takedown notice to our designated agent.
              </p>
              
              <h3 className="text-lg font-semibold mb-2 text-foreground">Requirements for a Valid Takedown Notice</h3>
              <p className="text-muted-foreground mb-3">
                Under 17 U.S.C. § 512(c)(3), a valid DMCA takedown notice must include <strong>all</strong> of 
                the following elements:
              </p>
              <ol className="list-decimal list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong className="text-foreground">Physical or Electronic Signature</strong> - Your handwritten signature (scanned) or electronic signature</li>
                <li><strong className="text-foreground">Identification of Copyrighted Work</strong> - Clear description of the work you claim has been infringed</li>
                <li><strong className="text-foreground">Identification of Infringing Material</strong> - Specific URL(s) on Clipper where the allegedly infringing material is located</li>
                <li><strong className="text-foreground">Your Contact Information</strong> - Full legal name, physical address, telephone number, and email address</li>
                <li><strong className="text-foreground">Good Faith Statement</strong> - A statement that you have a good faith belief that use is not authorized</li>
                <li><strong className="text-foreground">Accuracy Statement Under Penalty of Perjury</strong> - A statement that the information is accurate</li>
                <li><strong className="text-foreground">Authorization Statement</strong> - A statement that you are authorized to act on behalf of the copyright owner</li>
              </ol>

              <div className="mt-6 p-4 bg-muted/30 rounded-lg border border-border">
                <h4 className="font-semibold mb-2 text-foreground">How to Submit</h4>
                <p className="text-muted-foreground mb-2">
                  <strong>Preferred Method:</strong> Email to{' '}
                  <a href="mailto:dmca@clpr.tv" className="text-primary hover:underline">dmca@clpr.tv</a>
                </p>
                <p className="text-muted-foreground text-sm">
                  Subject line: "DMCA Takedown Notice"<br />
                  Include all required information in the email body or attached document
                </p>
              </div>
            </CardBody>
          </Card>

          {/* Our Response Process */}
          <Card id="response-process">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Our Response Process</h2>
              <p className="text-muted-foreground mb-4">
                Upon receiving a valid DMCA takedown notice, we will:
              </p>
              
              <div className="space-y-4">
                <div>
                  <h3 className="text-lg font-semibold mb-2 text-foreground">1. Initial Review (Within 24 Hours)</h3>
                  <p className="text-muted-foreground">
                    Verify the notice contains all required elements and request missing information if incomplete.
                  </p>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-2 text-foreground">2. Takedown Decision (Within 24-48 Hours)</h3>
                  <p className="text-muted-foreground">
                    If the notice is valid and complete, we will remove or disable access to the allegedly 
                    infringing material and document the takedown in our systems.
                  </p>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-2 text-foreground">3. User Notification (Within 48-72 Hours)</h3>
                  <p className="text-muted-foreground">
                    Notify the user who submitted the content via email, provide a copy of the DMCA notice, 
                    and explain their right to file a counter-notice.
                  </p>
                </div>
              </div>

              <div className="mt-4 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                <p className="text-sm text-muted-foreground">
                  <strong className="text-foreground">Important:</strong> We do not make legal judgments about 
                  whether content is actually infringing. We rely on the representations in your notice and act 
                  as a neutral intermediary between copyright holders and users.
                </p>
              </div>
            </CardBody>
          </Card>

          {/* Filing a Counter-Notice */}
          <Card id="counter-notice">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Filing a Counter-Notice</h2>
              <p className="text-muted-foreground mb-4">
                If your content was removed due to a DMCA takedown notice and you believe the removal was 
                a mistake or misidentification, you may file a counter-notice.
              </p>
              
              <h3 className="text-lg font-semibold mb-2 text-foreground">Requirements for a Valid Counter-Notice</h3>
              <p className="text-muted-foreground mb-3">
                Under 17 U.S.C. § 512(g)(3), a valid DMCA counter-notice must include:
              </p>
              <ol className="list-decimal list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong className="text-foreground">Physical or Electronic Signature</strong></li>
                <li><strong className="text-foreground">Identification of Removed Material</strong> - URL(s) and description of the material before removal</li>
                <li><strong className="text-foreground">Good Faith Statement</strong> - Under penalty of perjury that material was removed by mistake or misidentification</li>
                <li><strong className="text-foreground">Consent to Jurisdiction</strong> - You consent to Federal District Court jurisdiction</li>
                <li><strong className="text-foreground">Consent to Service of Process</strong> - You will accept service of process from the original complainant</li>
                <li><strong className="text-foreground">Your Contact Information</strong> - Full legal name, physical address, phone number, and email</li>
              </ol>

              <h3 className="text-lg font-semibold mb-2 text-foreground mt-6">Counter-Notice Response</h3>
              <p className="text-muted-foreground mb-3">
                Upon receiving a valid counter-notice:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li>We forward the counter-notice to the original copyright holder</li>
                <li>They have 10-14 business days to file a lawsuit</li>
                <li>If no lawsuit is filed, we restore the removed content</li>
                <li>If a lawsuit is filed, content remains removed pending court resolution</li>
              </ul>

              <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
                <p className="text-sm text-muted-foreground">
                  <strong className="text-foreground">Warning:</strong> Filing a counter-notice exposes you to 
                  potential litigation. The original complainant may sue you. Consult an attorney before filing 
                  a counter-notice if you're unsure about your rights.
                </p>
              </div>
            </CardBody>
          </Card>

          {/* Repeat Infringer Policy */}
          <Card id="repeat-infringer">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Repeat Infringer Policy</h2>
              <p className="text-muted-foreground mb-4">
                Clipper has adopted a policy of terminating, in appropriate circumstances, users who are 
                repeat infringers of copyright.
              </p>
              
              <h3 className="text-lg font-semibold mb-3 text-foreground">Three-Strike System</h3>
              <div className="space-y-3">
                <div className="p-3 bg-muted/30 rounded-lg">
                  <h4 className="font-semibold text-foreground mb-1">Strike 1 - First Offense</h4>
                  <p className="text-sm text-muted-foreground">
                    Content removed immediately. Warning sent via email. Account remains active with full access.
                  </p>
                </div>

                <div className="p-3 bg-muted/30 rounded-lg">
                  <h4 className="font-semibold text-foreground mb-1">Strike 2 - Second Offense</h4>
                  <p className="text-sm text-muted-foreground">
                    Content removed immediately. Temporary suspension of account for 7 days. Final warning issued.
                  </p>
                </div>

                <div className="p-3 bg-muted/30 rounded-lg">
                  <h4 className="font-semibold text-foreground mb-1">Strike 3 - Third Offense</h4>
                  <p className="text-sm text-muted-foreground">
                    Content removed immediately. Permanent account termination. User banned from creating new accounts.
                  </p>
                </div>
              </div>

              <div className="mt-4 space-y-2">
                <p className="text-muted-foreground">
                  <strong className="text-foreground">Strike Expiration:</strong> Strikes expire after 12 months 
                  from the date of the infringement notice. Users can rebuild their standing over time with 
                  compliant behavior.
                </p>
                <p className="text-muted-foreground">
                  <strong className="text-foreground">Appeal Process:</strong> Users may appeal strikes by 
                  submitting a written appeal to dmca@clpr.tv with evidence that the DMCA notice was invalid, 
                  they have permission, or the content qualifies as fair use.
                </p>
              </div>
            </CardBody>
          </Card>

          {/* False Claims */}
          <Card id="false-claims">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">False Claims and Misrepresentation</h2>
              <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg mb-4">
                <p className="text-foreground font-semibold mb-2 flex items-center gap-1"><AlertTriangle size={16} strokeWidth={1.75} /> Warning</p>
                <p className="text-muted-foreground text-sm">
                  Making false claims in a DMCA notice or counter-notice can result in serious legal consequences 
                  under 17 U.S.C. § 512(f), including damages, costs, attorneys' fees, and potential criminal 
                  charges for perjury.
                </p>
              </div>

              <p className="text-muted-foreground mb-3">
                If you submit a false or bad faith DMCA notice or counter-notice, you may face:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong className="text-foreground">Legal Liability:</strong> Damages, costs, and attorneys' fees</li>
                <li><strong className="text-foreground">Account Actions:</strong> Account suspension or termination</li>
                <li><strong className="text-foreground">Legal Action:</strong> We may pursue legal action against repeat abusers</li>
              </ul>

              <p className="text-muted-foreground mt-4">
                Before submitting a DMCA notice, verify you own the copyright, confirm the use is not authorized, 
                review the content carefully, and consider fair use defenses. Consult an attorney if you're 
                unsure about your rights.
              </p>
            </CardBody>
          </Card>

          {/* Fair Use */}
          <Card id="fair-use">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Fair Use and User Rights</h2>
              <p className="text-muted-foreground mb-4">
                "Fair use" is a legal doctrine (17 U.S.C. § 107) that permits limited use of copyrighted 
                material without permission for purposes such as commentary, criticism, news reporting, 
                teaching, scholarship, research, and parody.
              </p>
              
              <h3 className="text-lg font-semibold mb-2 text-foreground">Fair Use Factors</h3>
              <p className="text-muted-foreground mb-3">Courts consider four factors:</p>
              <ol className="list-decimal list-inside space-y-2 text-muted-foreground ml-4">
                <li>Purpose and character of use (commercial vs. non-profit educational; transformative nature)</li>
                <li>Nature of the copyrighted work (factual vs. creative)</li>
                <li>Amount and substantiality used in relation to the whole work</li>
                <li>Effect on the market for the original work</li>
              </ol>

              <div className="mt-4 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                <p className="text-sm text-muted-foreground">
                  <strong className="text-foreground">Important:</strong> Fair use is a legal defense, not a 
                  safe harbor. Clipper cannot make legal determinations about fair use. If you receive a DMCA 
                  notice and believe your use is fair use, consult an attorney and consider filing a counter-notice.
                </p>
              </div>
            </CardBody>
          </Card>

          {/* No Monitoring Obligation */}
          <Card id="no-monitoring">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">No Obligation to Monitor</h2>
              <p className="text-muted-foreground mb-4">
                Under 17 U.S.C. § 512(m), Clipper is not obligated to monitor user-submitted content for 
                copyright infringement. We do not actively monitor the Service for infringing content or 
                affirmatively seek facts indicating infringing activity.
              </p>
              <p className="text-muted-foreground">
                However, we respond promptly to valid DMCA notices, investigate reports from copyright holders, 
                and cooperate with law enforcement and legal authorities.
              </p>
            </CardBody>
          </Card>

          {/* Contact */}
          <Card id="contact">
            <CardBody>
              <h2 className="text-2xl font-semibold mb-4">Contact Information</h2>
              <p className="text-muted-foreground mb-4">
                For DMCA notices, counter-notices, and copyright-related inquiries:
              </p>
              <div className="space-y-2 text-muted-foreground ml-4">
                <p>
                  <strong className="text-foreground">DMCA Agent Email:</strong>{' '}
                  <a href="mailto:dmca@clpr.tv" className="text-primary hover:underline">
                    dmca@clpr.tv
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">Legal Inquiries:</strong>{' '}
                  <a href="mailto:legal@clpr.tv" className="text-primary hover:underline">
                    legal@clpr.tv
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">General Support:</strong>{' '}
                  <a href="mailto:support@clpr.tv" className="text-primary hover:underline">
                    support@clpr.tv
                  </a>
                </p>
              </div>

              <div className="mt-6 p-4 bg-muted/30 rounded-lg">
                <h3 className="font-semibold mb-2 text-foreground">Additional Resources</h3>
                <ul className="space-y-1 text-sm text-muted-foreground">
                  <li>
                    <a
                      href="https://www.copyright.gov/"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      U.S. Copyright Office
                    </a>
                  </li>
                  <li>
                    <a
                      href="https://www.law.cornell.edu/uscode/text/17/512"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      DMCA Text (17 U.S.C. § 512)
                    </a>
                  </li>
                  <li>
                    <a
                      href="https://www.copyright.gov/fair-use/"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      Fair Use Information
                    </a>
                  </li>
                </ul>
              </div>
            </CardBody>
          </Card>
        </div>
      </Container>
    </>
  );
}
