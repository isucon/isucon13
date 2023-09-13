import { CssVarsProvider } from '@mui/joy/styles';
import React, { Suspense } from 'react';
import { useRoutes } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import routes from '~react-pages';
import 'focus-visible/dist/focus-visible';

export function App(): React.ReactElement {
  const routeContent = useRoutes(routes);

  return (
    <CssVarsProvider defaultMode="light">
      <Layout>
        <Suspense fallback={<p>Loading...</p>}>{routeContent}</Suspense>
      </Layout>
    </CssVarsProvider>
  );
}
