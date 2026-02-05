# 0002. Adopt Rclone RC API for plat-rclone GUI

*   **Status**: Proposed
*   **Date**: 2026-02-05
*   **Authors**: Joeblew999, Gemini
*   **Reviewers**: Joeblew999

## Context

The `plat-rclone` project aims to provide a robust and feature-rich web GUI for `rclone`. The core functionality of `rclone` is exposed through its Remote Control (RC) API. The existing official `rclone` web GUI (`rclone-webui-react`), a ReactJS application, heavily relies on this RC API for all its operations, communicating with the `rclone` daemon typically running on `http://localhost:5572`.

To ensure `plat-rclone` achieves feature parity with, and potentially surpasses, the existing web GUI, it is crucial to adopt and fully integrate with `rclone`'s RC API. This approach will leverage the comprehensive and well-defined set of operations already provided by `rclone`.

The project is developed in Golang, and the web GUI will utilize `VIA 7 DATASTAR`. A significant consideration is the need to manage `rclone` installations across various operating systems, implying that `plat-rclone` will interact with the `rclone` binary regardless of the host OS.

Key `rclone` RC API endpoints identified from the existing web GUI include (but are not limited to):

*   `operations/mkdir`
*   `operations/purge`
*   `operations/deletefile`
*   `operations/publiclink`
*   `core/stats`
*   `core/bwlimit`
*   `sync/move`
*   `operations/movefile`
*   `sync/copy`
*   `operations/copyfile`
*   `operations/cleanup`
*   `rc/noopauth`
*   `core/version`
*   `core/memstats`
*   `options/get`
*   `config/providers`
*   `config/dump`
*   `job/list`
*   `job/status`
*   `config/get`
*   `config/create`
*   `config/update`
*   `operations/fsinfo`
*   `config/listremotes`
*   `operations/list`
*   `operations/about`
*   `config/delete`
*   `job/stop`
*   `mount/listmounts`
*   `mount/mount`
    *   `mount/unmount`
    *   `mount/unmountall`

## Consequences

Adopting the `rclone` RC API as the primary communication mechanism for `plat-rclone` offers several advantages:

*   **Feature Completeness:** Ensures immediate access to the full range of `rclone` functionalities, facilitating feature parity with the official web GUI and rapid development of new features.
*   **Reduced Development Overhead:** Avoids reimplementing complex file system operations and `rclone` core logic, as these are handled by the `rclone` binary itself.
*   **Consistency:** Maintains consistency with `rclone`'s behavior and data models, as the API is the official interface.
*   **Cross-OS Compatibility:** The `rclone` RC API is platform-agnostic, allowing `plat-rclone` to interact with `rclone` seamlessly across different operating systems, aligning with the project's goal of supporting all OS.

Potential drawbacks include:

*   **Dependency on Rclone Binary:** `plat-rclone` will be inherently dependent on the `rclone` binary being present and accessible on the system.
*   **API Evolution:** Future changes to `rclone`'s RC API might require updates to `plat-rclone` to maintain compatibility.
*   **Performance Considerations:** While `rclone` RC is efficient, careful handling of API calls and potential asynchronous processing within the Golang backend will be necessary to ensure a responsive user experience, especially when dealing with large datasets or numerous concurrent operations.

This decision firmly establishes the `rclone` RC API as the foundational interface for `plat-rclone`'s backend logic.
