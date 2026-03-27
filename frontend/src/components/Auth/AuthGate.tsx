"use client";

import { createContext, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import AuthChecker from "@/api/auth/authchecker";
import { PermedUser } from "@/api/permissions/permshelpers";
import { ProfilePictureProvider } from "@/components/Auth/ProfilePictureContext";

export const UserContext = createContext<PermedUser | undefined>(undefined);

export default function AuthGate({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();

    const [checked, setChecked] = useState(false);
    const [user, setUser] = useState<PermedUser | undefined>(undefined);

    useEffect(() => {
        if (pathname?.startsWith("/login")) {
            setChecked(true);
            return;
        }

        AuthChecker().then((ok) => {
            if (ok.some) {
                setChecked(true);
                setUser(ok.value);
            } else {
                router.replace("/login");
            }
        });
    }, [pathname, router]);

    if (!checked) return null;
    if (!user) return null;

    return (
        <UserContext value={user}>
            <ProfilePictureProvider>{children}</ProfilePictureProvider>
        </UserContext>
    );
}
