'use client';

import { Button } from '@/components/ui/button';
import Image from 'next/image';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import Link from 'next/link';
import { useContext } from 'react';
import { UserContext } from '../user_wrapper';

export function User() {
  const { user, logout } = useContext(UserContext);

  const isKeycloakProvider =
    process.env.NEXT_PUBLIC_IDENTITY_PROVIDER &&
    process.env.NEXT_PUBLIC_IDENTITY_PROVIDER.toLowerCase().includes('keycloak');

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="icon" className="overflow-hidden rounded-full">
          <Image
            src={'/placeholder-user.jpg'}
            width={36}
            height={36}
            alt="Avatar"
            className="overflow-hidden rounded-full"
          />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>{user?.username}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {/* <DropdownMenuItem>Settings</DropdownMenuItem>
        <DropdownMenuItem>Support</DropdownMenuItem> */}
        <DropdownMenuSeparator />
        {user && (
          <DropdownMenuItem asChild>
            <button
              onClick={(e) => {
                e.preventDefault();
                logout();
              }}
              className="w-full text-left"
            >
              Sign Out
            </button>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
