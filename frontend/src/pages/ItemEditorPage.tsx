import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api } from "../lib/api";
import type { ItemPayload, PickupPoint } from "../types/api";

interface ItemEditorPageProps {
  mode: "create" | "edit";
}

const emptyForm: ItemPayload = {
  pickup_point_id: "",
  title: "",
  description: "",
  category: "",
  price_per_hour: 0,
  price_per_day: 0,
};

export function ItemEditorPage({ mode }: ItemEditorPageProps) {
  const { itemId = "" } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState<ItemPayload>(emptyForm);
  const [initialForm, setInitialForm] = useState<ItemPayload>(emptyForm);
  const [pickupPoints, setPickupPoints] = useState<PickupPoint[]>([]);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    void api.getPickupPoints().then(setPickupPoints).catch(() => setPickupPoints([]));
  }, []);

  useEffect(() => {
    if (mode !== "edit") {
      return;
    }

    void Promise.all([api.getItem(itemId), api.getItems()])
      .then(([details, items]) => {
        const summary = items.find((item) => item.id === itemId);
        const nextForm = {
          pickup_point_id: details.pickup_point?.id ?? details.PickupPoint?.id ?? "",
          title: details.title ?? summary?.title ?? "",
          description: details.description ?? "",
          category: details.category ?? summary?.category ?? "",
          price_per_hour: details.price_per_hour || summary?.price_per_hour || 0,
          price_per_day: details.price_per_day || summary?.price_per_day || 0,
        };
        setForm(nextForm);
        setInitialForm(nextForm);
      })
      .catch((requestError: Error) => setError(requestError.message));
  }, [itemId, mode]);

  function updateField<K extends keyof ItemPayload>(field: K, value: ItemPayload[K]) {
    setForm((current) => ({
      ...current,
      [field]: value,
    }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    setMessage("");

    try {
      if (!form.pickup_point_id) {
        throw new Error("Please select a pickup point.");
      }

      if (mode === "create") {
        const createdItem = await api.createItem(form);
        if (uploadFile) {
          await api.uploadItemImage(createdItem.id, uploadFile);
        }
        navigate(`/items/${createdItem.id}`);
        return;
      }

      const patch: Partial<ItemPayload> = {};

      if (form.pickup_point_id !== initialForm.pickup_point_id) {
        patch.pickup_point_id = form.pickup_point_id;
      }
      if (form.title !== initialForm.title) {
        patch.title = form.title;
      }
      if (form.description !== initialForm.description) {
        patch.description = form.description;
      }
      if (form.category !== initialForm.category) {
        patch.category = form.category;
      }
      if (form.price_per_hour !== initialForm.price_per_hour) {
        patch.price_per_hour = form.price_per_hour;
      }
      if (form.price_per_day !== initialForm.price_per_day) {
        patch.price_per_day = form.price_per_day;
      }

      if (Object.keys(patch).length > 0) {
        await api.updateItem(itemId, patch);
        setMessage("Listing updated.");
      } else {
        setMessage("No listing fields changed.");
      }

      if (uploadFile) {
        await api.uploadItemImage(itemId, uploadFile);
        setMessage("Listing updated and image uploaded.");
      }
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to save item.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="stack-page">
      <section className="panel-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">{mode === "create" ? "New listing" : "Edit listing"}</p>
            <h1>{mode === "create" ? "Create an item for rent" : "Update your listing"}</h1>
          </div>
          <Link className="button secondary" to="/dashboard">
            Back to dashboard
          </Link>
        </div>

        <form className="editor-layout" onSubmit={handleSubmit}>
          <div className="panel-subsection">
            <label>
              <span>Title</span>
              <input value={form.title} onChange={(event) => updateField("title", event.target.value)} required />
            </label>
            <label>
              <span>Description</span>
              <textarea
                rows={6}
                value={form.description}
                onChange={(event) => updateField("description", event.target.value)}
              />
            </label>
            <label>
              <span>Category</span>
              <input value={form.category} onChange={(event) => updateField("category", event.target.value)} required />
            </label>
          </div>

          <div className="panel-subsection">
            <label>
              <span>Pickup point</span>
              <select
                value={form.pickup_point_id}
                onChange={(event) => updateField("pickup_point_id", event.target.value)}
                required
              >
                <option value="">Select pickup point</option>
                {pickupPoints.map((pickupPoint) => (
                  <option key={pickupPoint.id} value={pickupPoint.id}>
                    {pickupPoint.city} - {pickupPoint.address}
                  </option>
                ))}
              </select>
            </label>
            <p className="helper-text">Choose where the renter will collect the item.</p>
            <label>
              <span>Price per hour</span>
              <input
                type="number"
                min={0}
                value={form.price_per_hour}
                onChange={(event) => updateField("price_per_hour", Number(event.target.value))}
              />
            </label>
            <label>
              <span>Price per day</span>
              <input
                type="number"
                min={0}
                value={form.price_per_day}
                onChange={(event) => updateField("price_per_day", Number(event.target.value))}
              />
            </label>
            <label>
              <span>Upload image</span>
              <input type="file" accept="image/*" onChange={(event) => setUploadFile(event.target.files?.[0] ?? null)} />
            </label>
          </div>

          <div className="editor-footer">
            <button className="button" disabled={saving} type="submit">
              {saving ? "Saving..." : mode === "create" ? "Create listing" : "Save changes"}
            </button>
            {message ? <p className="helper-text success">{message}</p> : null}
            {error ? <p className="form-error">{error}</p> : null}
          </div>
        </form>
      </section>
    </div>
  );
}
