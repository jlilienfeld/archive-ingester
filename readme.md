# Overview

This microservice offers a solution for handling big archive files. It
decompresses and unpacks them upon upload, unpacking and extracting nested
archives within the main file. Its primary goal is to simplify researchers'
management of large archive files by simultaneously decoding,
storing, and cataloging their contents. While upload times may be prolonged due
to file size, the advantage lies in the likelihood that the comprehensive
recursive unwrapping and indexing process will conclude by the end of data
transfer.