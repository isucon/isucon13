import React from 'react';

interface LayoutProps {
  children: React.ReactNode;
}
export function Layout(props: LayoutProps): React.ReactElement {
  return <div>{props.children}</div>;
}
