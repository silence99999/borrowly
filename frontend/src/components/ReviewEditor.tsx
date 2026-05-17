import { useState } from "react";
import type { Review } from "../types/api";

interface ReviewEditorProps {
  review?: Review;
  onSubmit: (payload: { rating: number; comment: string }) => Promise<void>;
  onCancel?: () => void;
}

export function ReviewEditor({ review, onSubmit, onCancel }: ReviewEditorProps) {
  const [rating, setRating] = useState(review?.rating ?? 5);
  const [comment, setComment] = useState(review?.comment ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");

    try {
      await onSubmit({ rating, comment });
      if (!review) {
        setComment("");
        setRating(5);
      }
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Failed to save review.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="stacked-form" onSubmit={handleSubmit}>
      <label>
        <span>Rating</span>
        <input
          type="number"
          min={0}
          max={5}
          value={rating}
          onChange={(event) => setRating(Number(event.target.value))}
          required
        />
      </label>
      <label>
        <span>Comment</span>
        <textarea
          rows={4}
          value={comment}
          onChange={(event) => setComment(event.target.value)}
          placeholder="Share how the rental went."
        />
      </label>
      {error ? <p className="form-error">{error}</p> : null}
      <div className="inline-actions">
        <button className="button" disabled={saving} type="submit">
          {saving ? "Saving..." : review ? "Update review" : "Post review"}
        </button>
        {review && onCancel ? (
          <button className="button secondary" type="button" onClick={onCancel}>
            Cancel
          </button>
        ) : null}
      </div>
    </form>
  );
}
