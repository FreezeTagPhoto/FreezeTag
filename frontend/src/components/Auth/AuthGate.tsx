"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import AuthChecker from "@/api/auth/authchecker";

export default function AuthGate({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();

    const [checked, setChecked] = useState(false);
    const [authed, setAuthed] = useState(false);

    useEffect(() => {
        if (pathname?.startsWith("/login")) {
            setChecked(true);
            setAuthed(true);
            return;
        }

        AuthChecker().then((ok) => {
            setAuthed(ok);
            setChecked(true);
            if (!ok) {
                router.replace("/login");
            }
        });
    }, [pathname, router]);

    if (!checked) return null;
    if (!authed) return null;

    return <>{children}</>;
}
