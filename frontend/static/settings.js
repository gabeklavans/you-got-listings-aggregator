const brokers = [];

const FilterKeys = [
  "BedsMin",
  "BedsMax",
  "BathsMin",
  "BathsMax",
  "RentMin",
  "RentMax",
  "DateMin",
  "DateMax",
  "ExcludedAreas",
]

const NotificationKeys = [
  "NotificationUrls"
]

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

function displayBrokers() {
  const brokersListDiv = document.getElementById("brokers-list");

  // clear all current list items
  while (brokersListDiv.lastChild) {
    brokersListDiv.removeChild(brokersListDiv.lastChild);
  }

  for (let idx = 0; idx < brokers.length; idx++) {
    const broker = brokers[idx];

    $("#brokers-list").append(`
      <div id="broker-item-${idx}" class="broker-item">
        <button type="button" onclick="handleRemoveBroker(this)">Remove</button>
        <p>${broker.name}: ${broker.url}</p>
      </div>
    `);
  }
}

function handleAddBroker() {
  const nameElement = document.getElementById("broker-name");
  const urlElement = document.getElementById("broker-url");
  const name = nameElement.value
  const url = urlElement.value

  if (!name || !url) {
    return;
  }

  brokers.push({
    name,
    url,
  });

  console.debug("cur brokers:");
  console.debug(brokers);

  nameElement.value = "";
  urlElement.value = "";

  displayBrokers();
}

function handleRemoveBroker(removeButtonElement) {
  const brokerIdx = parseInt(removeButtonElement.parentElement.id.split("-")[2]);
  brokers.splice(brokerIdx, 1);

  displayBrokers();
}

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
    let notifications = []
    for (const [key, value] of formData) {
      if (!value) {
        continue;
      }
      console.log(key)

      if (FilterKeys.includes(key)) {
        filters.push({
          name: key,
          value: value,
        });
      } else if (NotificationKeys.includes(key)) {
        // split the notification URIs by comma into the proper JSON object structure
        notifications = value.split(",").map((notifUri) => {
          return {
            url: notifUri
          };
        });
      }
    }

    console.debug("sending brokers");
    console.debug(brokers)
    await fetch(`${window.location.origin}/v1/brokers`, {
      method: "PATCH",
      body: JSON.stringify(brokers)
    });

    console.debug("sending filters");
    console.debug(filters)
    await fetch(`${window.location.origin}/v1/filters`, {
      method: "PATCH",
      body: JSON.stringify(filters)
    });

    console.debug("sending notifications");
    console.debug(notifications)
    await fetch(`${window.location.origin}/v1/notifications`, {
      method: "PATCH",
      body: JSON.stringify(notifications)
    });

    alert(`I guess I'll filter on those.... not because I want to or anything...... >.<'`);

    document.location.assign(document.location.origin);
  },
  false,
);
