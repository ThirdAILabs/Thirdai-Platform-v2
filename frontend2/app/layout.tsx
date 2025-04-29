import './globals.css';

export const metadata = {
  title: 'ThirdAI Platform',
  description: 'Democratize AI for everyone.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="flex min-h-screen w-full flex-col bg-muted/40">
        {children}
      </body>
    </html>
  );
}
