# Assumptions

- There were mentions that people in the same group can sit next
to each other, but none about where to get groups information. So
I added a new APIs `/api/v1/groups` to query the groups.

- It doesn't state when the service is down, the room state is persisted.
Therefore, I didn't include a database in this project.

- The service can serves concurrent change state requests to a single resource.
Thus, it is important that serialization is ensured across all operations.

# Technical decisions

- For me, the fewer dependencies the better. I also want to try
how far the stdlib can go when writing http service. It's surprisingly good
and I think people should give it more chance.

- I try to keep things as simple as possible, with maintainability come first in mind.
I could optimize the query/reserve/cancel seats operation for large instance using parallel computing technique
due to the fact that the distance can be computed independently.

- Why the `reservedSeat` field of `RoomManager` is has the type of `map[in64]string`
but not `map[int]string`? it's because of the formula $x \times numCols + y$
might give us an integer overflow.