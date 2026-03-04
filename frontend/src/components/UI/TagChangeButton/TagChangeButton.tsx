import styles from "./TagChangeButton.module.css";
import { useEffect, useRef, useState } from "react";
import TagGetter from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";
import IndeterminateCheckbox, {
    CheckboxState,
} from "../IndeterminateCheckbox/IndeterminateCheckbox";
import TagIdCounter from "@/api/tags/tagidcounter";
import TagRemover from "@/api/tags/tagremover";

export type TagChangeProps = {
    image_ids: Set<number>;
    // set to false for new images, because they always have no seed other than unchecked
    do_seeding?: boolean;
};

// This makes a fake empty set because react hates empty set states for some reason
export function TagChangeEmptySet(): Set<number> {
    return new Set([0]);
}

export default function TagChangeButton({
    image_ids,
    do_seeding,
}: TagChangeProps) {
    const [allTags, setAllTags] = useState<string[]>([]);
    const [filteredTags, setFilteredTags] = useState<string[]>([]);
    const [searchQuery, setSearchQuery] = useState<string>("");

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
        const image_ids_array = image_ids.values().toArray();
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
    };

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
        if (image_ids.size === 0) {
            return;
        }
        if (!do_seeding) {
            return;
        }
        TagIdCounter(image_ids.values().toArray()).then((result) => {
            if (result.ok) {
                const newSeeds = new Map();
                // This way we don't count the 0 no-op
                const image_count = image_ids.size - (image_ids.has(0) ? 1 : 0);
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
            <div className={styles.tag_menu}>
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
                    ></input>
                    <button onClick={addNewTag}>New</button>
                </div>

                <label htmlFor="tags-submit" className={styles.label}>
                    Add Tags!
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
