"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import logoUrl from "@/icons/freezetag+text.svg";
import styles from "./Sidebar.module.css";

const navItems = [
  { label: "Upload", href: "/upload" },
  { label: "Manage", href: "/manage" },
  { label: "Settings", href: "/settings" },
  { label: "Plugins", href: "/plugins" },
  { label: "Accounts", href: "/accounts" },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <div className={styles.wrap}>
      <div className={styles.brand}>
        <Link href="/" className={styles.logoWrap}>
          <Image
            src={logoUrl}
            alt="FreezeTag"
            fill
            priority
            sizes="240px"
            className={styles.logoImg}
          />
        </Link>
      </div>

      <nav className={styles.menu}>
        {navItems.map((item) => {
          const isActive =
            item.href === "/"
              ? pathname === "/"
              : pathname === item.href || pathname.startsWith(item.href + "/");

          return (
            <Link
              key={item.label}
              href={item.href}
              className={`${styles.item} ${isActive ? styles.itemActive : ""}`}
              aria-current={isActive ? "page" : undefined}
            >
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
