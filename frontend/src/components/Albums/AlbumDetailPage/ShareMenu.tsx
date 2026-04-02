import { useState, useEffect, useContext } from "react";
import {
    Share2,
    ChevronDown,
    Loader2,
    X,
    User as UserIcon,
} from "lucide-react";
import UserLister, { User } from "@/api/users/userlister";
import styles from "./AlbumDetailPage.module.css";
import ProfilePictureGetter from "@/api/users/profilepicturegetter";
import { UserContext } from "@/components/Auth/AuthGate";
import { UserHasPerm } from "@/api/permissions/permshelpers";

