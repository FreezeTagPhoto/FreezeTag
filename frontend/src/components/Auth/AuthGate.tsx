"use client";

import { createContext, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import AuthChecker, { User } from "@/api/auth/authchecker";

export const UserContext = createContext<User | undefined>(undefined);

export default function AuthGate({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();

    const [checked, setChecked] = useState(false);
    const [user, setUser] = useState<User | undefined>(undefined);

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

    return <UserContext value={user}>{children}</UserContext>;
}
