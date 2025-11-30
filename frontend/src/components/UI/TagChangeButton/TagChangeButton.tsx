import styles from "./TagChangeButton.module.css";
import { useEffect, useState } from "react";
import TagGetter, { TagGetResult } from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";

export type TagChangeProps = {
    image_ids: Set<number>;
};

export default function TagChangeButton(props: TagChangeProps) {
    const [tags, setTags] = useState<string[]>([]);

    const updateTags = () => {
        TagGetter()
            .then((result: TagGetResult) => {
                if (result.ok) {
                    setTags(result.value);
                } else {
                    console.error("Error retrieving tags:", result.error);
                    // TODO: show error to user
                }
            })
            .catch(
                (error) =>
                    console.error(
                        "Error retrieving tags (is the backend running?):",
                        error,
                    ),
                // TODO: Show error to user
            );
    };

    const handleSubmit = async (event: FormData, image_ids: Set<number>) => {
        try {
            const tags = event
                .getAll("tag_menu")
                .map((entry) => entry.toString());
            const new_tag = event.get("new_tag")?.toString();
            if ((tags.length === 0 && !new_tag) || image_ids.size === 0) {
                console.error(
                    "Must have at least some tags and some images selected!",
                );
            } else {
                if (new_tag) tags.push(new_tag);
                const image_id_array = image_ids.values().toArray();
                const result = await TagAdder(image_id_array, tags);

                if (result.ok) {
                    console.log(
                        "Successfully added tags! result: ",
                        result.value,
                    );
                } else {
                    console.error("Error adding tags!", result.error);
                }
            }
            updateTags();
        } catch (error) {
            console.error(
                "Error adding tags (is the backend running?):",
                error,
            );
            // TODO: show error to user
        }
    };

    useEffect(updateTags, []);

    return (
        <form
            action={(e) => handleSubmit(e, props.image_ids)}
            className={styles.form}
        >
            <select multiple name="tag_menu" className={styles.tag_menu}>
                {tags.map((tag) => (
                    <option value={tag} key={tag}>
                        {tag}
                    </option>
                ))}
            </select>
            <input
                name="new_tag"
                type="text"
                placeholder="New tag..."
                className={styles.new_tag}
            ></input>
            <label htmlFor="tags-submit" className={styles.label}>
                {" "}
                Submit Tags!{" "}
            </label>
            <input type="submit" id="tags-submit" className={styles.button} />
        </form>
    );
}
