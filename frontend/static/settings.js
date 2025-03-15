function minMaxBoundsCheck(minElement, maxElement) {
  const elements = [minElement, maxElement];
  for (const element of elements) {
    element.addEventListener("input", _ => {
      if (parseInt(minElement.value.replaceAll("-", "")) > parseInt(maxElement.value.replaceAll("-", ""))) {
        element.setCustomValidity("Maximum cannot be greater than minimum.")
      } else {
        element.setCustomValidity("")
      }
    });
  }
}

minMaxBoundsCheck(document.getElementsByName("BedsMin")[0], document.getElementsByName("BedsMax")[0]);
minMaxBoundsCheck(document.getElementsByName("BathsMin")[0], document.getElementsByName("BathsMax")[0]);
minMaxBoundsCheck(document.getElementsByName("RentMin")[0], document.getElementsByName("RentMax")[0]);
minMaxBoundsCheck(document.getElementsByName("DateMin")[0], document.getElementsByName("DateMax")[0]);

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

    console.debug("sending filters");
    console.debug(filters)
    await fetch(`${window.location.origin}/v1/filters`, {
      method: "PATCH",
      body: JSON.stringify(filters)
    });

    alert(`I guess I'll filter on those.... not because I want to or anything...... >.<'`);

    document.location.assign(document.location.origin);
  },
  false,
);
