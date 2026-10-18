import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import CodeBlock from '@theme/CodeBlock';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import styles from './index.module.css';

const installer = 'curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh';

export default function Home(): ReactNode {
  return (
    <Layout title="Quick start" description="Too busy to vote on AZs? Install AZ Vote Delegate, choose a pillar you trust, and automatically copy its proposal and phase-payout votes. Keep manual control.">
      <main>
        <header className={styles.hero}>
          <div className="container">
            <p className="ledger">Zenon Network of Momentum · Accelerator-Z voting</p>
            <h1>AZ Vote Delegate</h1>
            <p className={styles.lead}>Too busy to vote on AZs? Delegate to a pillar you trust.</p>
            <p className={styles.intro}>Install a lightweight app on your pillar, follow the setup menu, and automatically copy your chosen pillar’s votes on AZ proposals and phase payouts.</p>
            <p className={styles.intro}>Keep your wallet seed private. The app uses your existing producer key, and you can cast your own votes anytime.</p>
          </div>
        </header>
        <section className={`container ${styles.quickstart}`} aria-labelledby="quick-start">
          <div className={styles.sectionHeading}><h2 id="quick-start">Quick start</h2><p>Install. Set up. Let your pillar vote.</p></div>
          <p className={styles.requirements}>For Linux amd64 / arm64 pillar hosts with an existing producer key. Uses legacy fused plasma / proof of work; <strong>Dynamic Plasma is not supported.</strong></p>
          <article className={styles.install}>
            <div className={styles.stepTitle}><span className={styles.number}>01</span><div><p className="ledger">On your pillar host</p><Heading as="h3" id="install">Install and open setup</Heading></div></div>
            <p>Run as the user who owns your producer key. Select the matching command below.</p>
            <Tabs lazy groupId="operator" defaultValue="root" values={[{label: 'Node runs as root', value: 'root'}, {label: 'Node runs as a regular user', value: 'user'}]}>
              <TabItem value="root"><CodeBlock language="bash">{`${installer} | sudo sh`}</CodeBlock></TabItem>
              <TabItem value="user"><CodeBlock language="bash">{`${installer} | sh`}</CodeBlock></TabItem>
            </Tabs>
            <p className={styles.caption}>The installer downloads a verified binary and opens the terminal setup menu. No compiler needed. <Link to="/getting-started">Installation details →</Link></p>
          </article>
          <div className={styles.steps}>
            <article className={styles.step} id="setup">
              <div className={styles.stepTitle}><span className={styles.number}>02</span><div><p className="ledger">In the setup menu</p><h3>Connect and choose a pillar</h3></div></div>
              <ol>
                <li>Confirm your node URL and chain ID. Defaults: <code>ws://127.0.0.1:35998</code> and mainnet <code>1</code>.</li>
                <li>Choose <strong>Yes</strong> to install the background service.</li>
                <li>Select <strong>Use existing znnd configuration</strong>, or enter your producer key file and password.</li>
                <li>Choose a pillar to copy from the shuffled top 30 voters. Participation is shown beside each name.</li>
                <li>Set a <strong>copy vote delay</strong> in minutes, or leave it at <strong>0</strong> for no delay. Manual votes bypass it.</li>
              </ol>
              <Link to="/delegation">Delegate selection and vote delays →</Link>
            </article>
            <article className={styles.step} id="run">
              <div className={styles.stepTitle}><span className={styles.number}>03</span><div><p className="ledger">Choose how to run</p><h3>Preview or start voting</h3></div></div>
              <div className={styles.mode}><strong>Dry run</strong><p>Preview what the service would do. No votes are submitted. Selected by default.</p></div>
              <div className={styles.mode}><strong>Live</strong><p>Copy the selected pillar’s votes on AZ proposals and phase-payout requests.</p></div>
              <p>Confirm <strong>Save and start</strong>. You can close the menu; the service keeps running. Change modes anytime from <strong>Voting mode: Dry run / Live</strong>.</p>
              <Link to="/service">Check that the service is running →</Link>
            </article>
          </div>
          <article className={styles.reopen}>
            <div><p className="ledger">You’re in control</p><h3>Come back whenever you need to</h3><p>Change the pillar you copy, adjust the copy delay, cast a manual vote, or upgrade the app.</p></div>
            <div className={styles.reopenCommand}>
              <Tabs lazy groupId="operator" defaultValue="root" values={[{label: 'Root', value: 'root'}, {label: 'Regular user', value: 'user'}]}>
                <TabItem value="root"><CodeBlock language="bash">sudo az-vote-delegate</CodeBlock></TabItem>
                <TabItem value="user"><CodeBlock language="bash">~/.local/bin/az-vote-delegate</CodeBlock></TabItem>
              </Tabs>
              <p className={styles.caption}><kbd>Esc</kbd> goes back. Leaving the menu keeps the service running.</p>
            </div>
          </article>
          <div className={styles.more}>
            <Link className={styles.guide} to="/manual-voting"><p className="ledger">Manual control</p><h3>Your vote takes priority <span aria-hidden="true">→</span></h3><p>A manual vote protects that request. Other proposals and phases keep copying.</p></Link>
            <Link className={styles.guide} to="/updates"><p className="ledger">Stay current</p><h3>Upgrade from the menu <span aria-hidden="true">→</span></h3><p>Check for new releases and upgrade while preserving settings and manual overrides.</p></Link>
            <Link className={styles.guide} to="/troubleshooting"><p className="ledger">Need a hand?</p><h3>Find the next step <span aria-hidden="true">→</span></h3><p>Check service status, connection problems, missing requests and pending votes.</p></Link>
          </div>
        </section>
      </main>
    </Layout>
  );
}
