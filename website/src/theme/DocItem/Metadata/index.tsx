import type {ReactNode} from 'react';
import Head from '@docusaurus/Head';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {useDoc} from '@docusaurus/plugin-content-docs/client';
import OriginalMetadata from '@theme-original/DocItem/Metadata';

export default function DocMetadata(): ReactNode {
  const {metadata} = useDoc();
  const {siteConfig} = useDocusaurusContext();
  const permalink = siteConfig.trailingSlash
    ? `${metadata.permalink.replace(/\/$/, '')}/`
    : metadata.permalink;
  const url = new URL(permalink, siteConfig.url).href;
  const schema = {
    '@context': 'https://schema.org',
    '@type': 'TechArticle',
    headline: metadata.title,
    description: metadata.description,
    url,
    inLanguage: 'en',
    isPartOf: {'@id': `${siteConfig.url}${siteConfig.baseUrl}#website`},
  };
  return <>
    <OriginalMetadata />
    <Head>
      <link rel="alternate" type="text/markdown" href={`${url.replace(/\/$/, '')}.md`} title={`${metadata.title} (Markdown)`} />
      <script type="application/ld+json">{JSON.stringify(schema).replace(/</g, '\\u003c')}</script>
    </Head>
  </>;
}
