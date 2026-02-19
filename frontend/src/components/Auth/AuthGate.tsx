"use client";

import { createContext, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import AuthChecker from "@/api/auth/authchecker";

export const UserIdContext = createContext(-1);

export default function AuthGate({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();

    const [checked, setChecked] = useState(false);
    const [userId, setUserId] = useState(-1);

    useEffect(() => {
        if (pathname?.startsWith("/login")) {
            setChecked(true);
            return;
        }

        AuthChecker().then((ok) => {
            if (ok.some) {
                setChecked(true);
                setUserId(ok.value);
            } else {
                router.replace("/login");
            }
        });
    }, [pathname, router]);

    if (!checked) return null;
    if (userId === -1) return null;

    return <UserIdContext value={userId}>{children}</UserIdContext>;
}
