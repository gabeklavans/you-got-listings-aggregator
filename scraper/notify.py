import asyncio
import logging
from typing import List

import apprise

logger = logging.getLogger(__name__)

apobj = apprise.Apprise()

def register_notifications(notifs: List[str]):
    for notif in notifs:
        apobj.add(notif) 

def notify(url: str) -> None:
    asyncio.run(apobj.async_notify(
        title='A wild listing appeared!',
        body=url,
    ))
