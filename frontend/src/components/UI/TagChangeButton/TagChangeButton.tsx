import styles from "./TagChangeButton.module.css";
import { useEffect, useRef, useState } from "react";
import TagGetter from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";
import IndeterminateCheckbox, {
    CheckboxState,
} from "../IndeterminateCheckbox/IndeterminateCheckbox";
import TagIdCounter from "@/api/tags/tagidcounter";
import TagRemover from "@/api/tags/tagremover";
import FreezeTagSet from "@/common/freezetag/freezetagset";
import { CirclePlus, Tag } from "lucide-react";

export type TagChangeProps = {
    image_ids: FreezeTagSet<number>;
    // set to false for new images, because they always have no seed other than unchecked
    do_seeding?: boolean;
};

export default function TagChangeButton({
    image_ids,
    do_seeding,
}: TagChangeProps) {
    const [allTags, setAllTags] = useState<string[]>([]);
    const [filteredTags, setFilteredTags] = useState<string[]>([]);
    const [searchQuery, setSearchQuery] = useState<string>("");
    const [fade, setFade] = useState({ top: false, bottom: false });
    const tagScrollRef = useRef<HTMLDivElement | null>(null);

    const updateAllTags = async () => {
        const result = await TagGetter();
        if (result.ok) {
            setAllTags(result.value);
            setFilteredTags(result.value);
        } else {
            console.error("Error retrieving tags:", result.error);
            // TODO: show error to user
        }
    };

    const handleSubmit = async () => {
        const image_ids_array = image_ids.toArray();
        const added_tags: string[] = [];
        const removed_tags: string[] = [];

        changedCheckboxes.forEach((val, key) => {
            if (val === CheckboxState.Checked) {
                added_tags.push(key);
            } else {
                removed_tags.push(key);
            }
        });

        if (added_tags.length !== 0) {
            await TagAdder(image_ids_array, added_tags);
        }
        if (removed_tags.length !== 0) {
            await TagRemover(image_ids_array, removed_tags);
        }

        setChangedCheckboxes(new Map());
    };

    const syncFade = () => {
        const el = tagScrollRef.current;
        if (!el) {
            setFade((prev) =>
                prev.top === false && prev.bottom === false
                    ? prev
                    : { top: false, bottom: false },
            );
            return;
        }
        const overflow = el.scrollHeight > el.clientHeight + 1;
        if (!overflow) {
            setFade((prev) =>
                prev.top === false && prev.bottom === false
                    ? prev
                    : { top: false, bottom: false },
            );
            return;
        }
        const atTop = el.scrollTop <= 1;
        const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1;
        const next = { top: !atTop, bottom: !atBottom };
        setFade((prev) =>
            prev.top === next.top && prev.bottom === next.bottom ? prev : next,
        );
    };

    useEffect(() => {
        const el = tagScrollRef.current;
        if (!el) return;
        const raf = requestAnimationFrame(syncFade);
        const onResize = () => syncFade();
        window.addEventListener("resize", onResize);
        const ro = new ResizeObserver(() => syncFade());
        ro.observe(el);
        return () => {
            cancelAnimationFrame(raf);
            window.removeEventListener("resize", onResize);
            ro.disconnect();
        };
    }, [filteredTags]);

    const searchTagRef = useRef<HTMLInputElement | null>(null);

    const addNewTag = async () => {
        const tag = searchTagRef.current?.value;
        if (!tag || tag === "") {
            return;
        }

        await TagAdder([], [tag]);
        await updateAllTags();
        setCheckboxSeeds(checkboxSeeds.set(tag, CheckboxState.Unchecked));
    };

    useEffect(() => {
        const arr = allTags.filter((tag) => tag.includes(searchQuery));
        setFilteredTags(arr);
    }, [searchQuery, allTags]);

    const isButtonDisabled = () =>
        searchQuery === "" || !!filteredTags.find((tag) => tag === searchQuery);

    useEffect(() => {
        updateAllTags();
    }, []);

    const [changedCheckboxes, setChangedCheckboxes] = useState<
        Map<string, CheckboxState>
    >(new Map());
    const [checkboxSeeds, setCheckboxSeeds] = useState<
        Map<string, CheckboxState>
    >(new Map());

    useEffect(() => {
        if (!do_seeding) {
            return;
        }
        if (image_ids.size() === 0) {
            setCheckboxSeeds(new Map());
            return;
        }
        TagIdCounter(image_ids.toArray()).then((result) => {
            if (result.ok) {
                const newSeeds = new Map();

                const image_count = image_ids.size();
                Object.entries(result.value).forEach(([tag, count]) => {
                    if (count === image_count) {
                        newSeeds.set(tag, CheckboxState.Checked);
                    } else if (count === 0) {
                        newSeeds.set(tag, CheckboxState.Unchecked);
                    } else {
                        newSeeds.set(tag, CheckboxState.Indeterminate);
                    }
                });
                setCheckboxSeeds(newSeeds);
            } else {
                console.error("tag id count error");
            }
        });
    }, [image_ids, do_seeding]);

    return (
        <div className={styles.form}>
            <div
                className={styles.tag_menu_wrap}
                data-fade-top={fade.top ? "1" : "0"}
                data-fade-bottom={fade.bottom ? "1" : "0"}
            >
                <div
                    className={styles.tag_menu}
                    ref={tagScrollRef}
                    onScroll={syncFade}
                >
                    {filteredTags.map((tag) => (
                        <label key={tag} title={tag}>
                            <IndeterminateCheckbox
                                value={
                                    (changedCheckboxes.has(tag)
                                        ? changedCheckboxes.get(tag)
                                        : checkboxSeeds.get(tag)) ??
                                    CheckboxState.Unchecked
                                }
                                afterChange={(val) =>
                                    setChangedCheckboxes(
                                        changedCheckboxes.set(tag, val),
                                    )
                                }
                            />
                            <p>{tag}</p>
                        </label>
                    ))}
                </div>
            </div>

            <div className={styles.bottom_container}>
                <div className={styles.new_tag_container}>
                    <input
                        name="new_tag"
                        type="text"
                        placeholder="Search tags..."
                        className={styles.new_tag}
                        onChange={(event) => setSearchQuery(event.target.value)}
                        autoComplete="off"
                        ref={searchTagRef}
                        title="Searches for a tag, or allows you to create a new tag if there is no match"
                    ></input>
                    <button
                        onClick={addNewTag}
                        title="Creates the tag typed into the search bar"
                        disabled={isButtonDisabled()}
                    >
                        <CirclePlus className={styles.iconBtnIcon} />
                    </button>
                </div>

                <label
                    htmlFor="tags-submit"
                    className={styles.label}
                    title="Adds selected tags to the selected images"
                >
                    <Tag className={styles.labelIcon} aria-hidden="true" />
                    Add Selected Tags
                </label>
                <input
                    type="submit"
                    id="tags-submit"
                    onClick={handleSubmit}
                    className={styles.button}
                />
            </div>
        </div>
    );
}
