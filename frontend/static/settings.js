const LIST_TYPE = {
  BROKER: "broker",
  EXCLUDED_AREA: "excludearea",
  NOTIFICATION: "notification",
};

let brokers = [];
let excludedAreas = [];
let notifications = [];

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

async function populateCurrentSettings() {
  // fill in the broker and notifications buffer
  let response = await fetch(`${window.location.origin}/v1/brokers`, {
    method: "GET",
  });
  if (!response.ok) {
    throw new Error(`Broker fetch status :${response.status}`)
  }
  const brokersJson = await response.json();
  if (brokersJson) {
    brokers = brokersJson.map(item => Object.values(item));
  } else {
    brokers = [];
  }
  displayList(LIST_TYPE.BROKER);

  response = await fetch(`${window.location.origin}/v1/notifications`, {
    method: "GET",
  });
  if (!response.ok) {
    throw new Error(`Notfication fetch status :${response.status}`)
  }
  const notificationsJson = await response.json();
  if (notificationsJson) {
    notifications = notificationsJson.map(item => Object.values(item));
  } else {
    notifications = [];
  }
  displayList(LIST_TYPE.NOTIFICATION);

  // filters are handled differently
  response = await fetch(`${window.location.origin}/v1/filters`, {
    method: "GET",
  });
  if (!response.ok) {
    throw new Error(`Filter fetch status :${response.status}`)
  }
  const filters = await response.json();
  for (const filter of filters) {
    if (filter.name === "ExcludeAreas") {
      excludedAreas = filter.value.split(",").map(area => [area]) // this is to please the generic displayList function
    } else {
      const inputElement = $(`[name="${filter.name}"]`)[0];
      if (!inputElement) {
        continue;
      }
      inputElement.value = `${filter.value}`;
    }
  }
  displayList(LIST_TYPE.EXCLUDED_AREA)
}

function getList(listType) {
  switch (listType) {
    case LIST_TYPE.BROKER:
      return brokers;
    case LIST_TYPE.EXCLUDED_AREA:
      return excludedAreas;
    case LIST_TYPE.NOTIFICATION:
      return notifications;
  }
}

function displayList(listType) {
  const list = getList(listType);

  const listDiv = $(`#${listType}s-list`);

  // clear all current list items
  listDiv.children().remove()

  for (let idx = 0; idx < list.length; idx++) {
    const listItem = list[idx];

    listDiv.append(`
      <div id="${listType}-item-${idx}" class="list-item">
        <button type="button" onclick="handleRemoveListItem(this)">Remove</button>
        <p>${listItem.join(": ")}</p>
      </div>
    `);
  }
}

function handleAddListItem(listType) {
  const list = getList(listType);
  const elements = Array.from($(`.${listType}-input`));

  if (elements.some(element => !element.value)) {
    return;
  }

  list.push(elements.map(element => element.value));

  console.debug(`cur ${listType}s:`);
  console.debug(list);

  for (const element of elements) {
    element.value = "";
  }

  displayList(listType);
}

function handleRemoveListItem(removeButtonElement) {
  const listType = removeButtonElement.parentElement.id.split("-")[0]
  const brokerIdx = parseInt(removeButtonElement.parentElement.id.split("-")[2]);

  let list = getList(listType);
  list.splice(brokerIdx, 1);

  displayList(listType);
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

    let hadReqErr = false;

    const filters = []
    // NOTE: form data should only contain Filter values
    // the rest of the values are handled separately
    for (const [key, value] of formData) {
      if (!value) {
        continue;
      }

      filters.push({
        name: key,
        value: value,
      });
    }

    // transform into proper JSON
    brokers = brokers.map(broker => {
      return {
        name: broker[0],
        url: broker[1],
      }
    });
    console.debug("sending brokers");
    console.debug(brokers)
    try {
      await fetch(`${window.location.origin}/v1/brokers`, {
        method: "PATCH",
        body: JSON.stringify(brokers)
      });
    } catch (error) {
      alert("Something went wrong saving the broker settings...");
      hadReqErr = true;
    }

    // massage excluded areas into the filters first
    if (excludedAreas.length > 0) {
      filters.push({
        name: "ExcludeAreas",
        value: excludedAreas.flat().join(","),
      });
    }
    console.debug("sending filters");
    console.debug(filters)
    try {
      await fetch(`${window.location.origin}/v1/filters`, {
        method: "PATCH",
        body: JSON.stringify(filters)
      });
    } catch (error) {
      alert("Something went wrong saving the filter settings...");
      hadReqErr = true;
    }

    // transform into proper JSON
    notifications = notifications.map(notif => {
      return {
        url: notif[0]
      };
    });
    console.debug("sending notifications");
    console.debug(notifications)
    try {
      await fetch(`${window.location.origin}/v1/notifications`, {
        method: "PATCH",
        body: JSON.stringify(notifications)
      });
    } catch (error) {
      alert("Something went wrong saving the notification settings...");
      hadReqErr = true;
    }

    if (hadReqErr) {
      return;
    }

    alert(`I guess I'll save those settings.... not because I want to or anything...... >.<'`);

    document.location.assign(document.location.origin);
  },
  false,
);

$(window).ready(populateCurrentSettings);
