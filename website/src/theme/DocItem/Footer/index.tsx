import type {ReactNode} from 'react';
import {useDoc} from '@docusaurus/plugin-content-docs/client';
import OriginalFooter from '@theme-original/DocItem/Footer';

export default function DocFooter(): ReactNode {
  const {metadata} = useDoc();
  return <>
    <p className="margin-top--lg"><a href={`${metadata.permalink.replace(/\/$/, '')}.md`}>Read this page as Markdown</a></p>
    <OriginalFooter />
  </>;
}
