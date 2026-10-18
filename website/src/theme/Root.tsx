import type {ReactNode} from 'react';
import Head from '@docusaurus/Head';
import {useLocation} from '@docusaurus/router';

export default function Root({children}: {children: ReactNode}): ReactNode {
  const {pathname} = useLocation();
  const utilityPage = /\/(search\/?|404\.html)$/.test(pathname);
  return <>
    <Head><meta name="robots" content={utilityPage ? 'noindex, follow' : 'index, follow, max-image-preview:large'} /></Head>
    {children}
  </>;
}
