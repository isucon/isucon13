/* eslint-disable @typescript-eslint/explicit-module-boundary-types */

import useSWR from 'swr';
import { apiClient } from './client';

export function useLiveStreams() {
  return useSWR('/livestream', () =>
    apiClient.get$livestream({
      parameter: {
        limit: 20,
      },
    }),
  );
}
