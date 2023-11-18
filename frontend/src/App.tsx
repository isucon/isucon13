import { CssVarsProvider } from '@mui/joy/styles';
import React, { Suspense } from 'react';
import { useRoutes } from 'react-router-dom';
import { SWRConfig } from 'swr';
import { HTTPError } from './api/client';
import { Layout } from './components/layout/Layout';
import { LoginModal } from './components/layout/loginmodal';
import routes from '~react-pages';
import 'focus-visible/dist/focus-visible';

export function App(): React.ReactElement {
  const routeContent = useRoutes(routes);
  const [isOpenReLoginModal, setIsOpenReLoginModal] = React.useState(false);

  return (
    <CssVarsProvider defaultMode="light">
      <SWRConfig
        value={{
          focusThrottleInterval: 20000,
          dedupingInterval: 10000,

          onError: (error) => {
            if (error instanceof HTTPError) {
              switch (error.response.status) {
                case 401:
                case 403:
                  setIsOpenReLoginModal(true);
                  break;
              }
            }
          },
        }}
      >
        <Layout>
          <LoginModal
            isOpen={isOpenReLoginModal}
            onClose={() => setIsOpenReLoginModal(false)}
          />
          <Suspense fallback={<p>Loading...</p>}>{routeContent}</Suspense>
        </Layout>
      </SWRConfig>
    </CssVarsProvider>
  );
}
