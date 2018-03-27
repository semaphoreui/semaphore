define(function () {
    app.registerController('ProjectInventoryCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
        $scope.reload = function () {
            $http.get(Project.getURL() + '/inventory?sort=name&order=asc').then(function (inventory) {
                $scope.inventory = inventory.data;
            });
        }

        $scope.remove = function (inventory) {
            $http.delete(Project.getURL() + '/inventory/' + inventory.id)
                .then(function () {
                    $scope.reload();
                })
                .catch(function (response) {
                    var d = response.data;
                    if (!(d && d.inUse)) {
                        swal('error', 'could not delete inventory..', 'error');
                        return;
                    }

                    swal({
                        title: 'Inventory in use',
                        text: d.error,
                        icon: 'error',
                        buttons: {
                            cancel: true,
                            confirm: {
                                text: 'Mark as removed',
                                closeModel: false,
                                className: 'bg-danger',
                            }
                        }
                    }).then(function (value) {
                        if (!value) {
                            return
                        }

                        $http
                            .delete(Project.getURL() + '/inventory/' + inventory.id + '?setRemoved=1')
                            .then(function () {
                                swal.stopLoading();
                                swal.close();

                                $scope.reload();
                            })
                            .catch(function () {
                                swal('Error', 'Could not delete inventory..', 'error');
                            });
                    })
                });
        }

        $scope.add = function () {
            $scope.getKeys(function (keys) {
                var scope = $rootScope.$new();
                scope.sshKeys = keys;

                $modal.open({
                    templateUrl: '/tpl/projects/inventory/add.html',
                    scope: scope
                }).result.then(function (inventory) {
                    $http.post(Project.getURL() + '/inventory', inventory.inventory)
                        .then(function () {
                            $scope.reload();
                        }).catch(function (response) {
                        swal('Error', 'Inventory not added: ' + response.status, 'error');
                    });
                }, function () {
                });
            });
        }

        $scope.edit = function (inventory) {
            $scope.getKeys(function (keys) {
                var scope = $rootScope.$new();
                scope.sshKeys = keys;
                scope.inventory = JSON.parse(JSON.stringify(inventory));

                $modal.open({
                    templateUrl: '/tpl/projects/inventory/add.html',
                    scope: scope
                }).result.then(function (opts) {
                    if (opts.remove) {
                        console.log(inventory)
                        return $scope.remove(inventory);
                    }

                    $http.put(Project.getURL() + '/inventory/' + inventory.id, opts.inventory)
                        .then(function () {
                            $scope.reload();
                        }).catch(function (response) {
                        swal('Error', 'Inventory not updated: ' + response.status, 'error');
                    });
                }, function () {
                });
            });
        }

        $scope.editContent = function (inventory) {
            var scope = $rootScope.$new();
            scope.inventory = inventory.inventory;

            $modal.open({
                templateUrl: '/tpl/projects/inventory/edit.html',
                scope: scope
            }).result.then(function (v) {
                inventory.inventory = v;
                $http.put(Project.getURL() + '/inventory/' + inventory.id, inventory)
                    .then(function () {
                        $scope.reload();
                    }).catch(function (response) {
                    swal('Error', 'Inventory not updated: ' + response.status, 'error');
                });
            }, function () {
            });
        }

        $scope.getKeys = function (cb) {
            if (typeof cb != 'function') cb = function () {
            }

            $http.get(Project.getURL() + '/keys?type=ssh').then(function (keys) {
                cb(keys.data);
            });
        }

        $scope.reload();
    }]);
});
