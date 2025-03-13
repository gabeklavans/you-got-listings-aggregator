settingsForm = document.forms["settings"];
settingsForm.addEventListener(
  "submit",
  async (event) => {
    event.preventDefault();
    const formData = new FormData(
      settingsForm,
      document.querySelector("button[value=Save]"),
    );

    const filters = []
    for (const [key, value] of formData) {
      if (!value) {
        continue;
      }

      filters.push({
        "name": key,
        "value": value,
      });
    }

    await fetch(`${window.location.origin}/v1/filters`, {
      method: "PATCH",
      body: JSON.stringify(filters)
    });

    alert(`I guess I'll filter on those.... not because I want to or anything...... >.<'`);

    document.location.assign(document.location.origin);
  },
  false,
);
